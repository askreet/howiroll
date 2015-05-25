package main

import (
	"amazonasg"
	"config"
	"fmt"
	"os"
	"strings"
	"waitfor"

	"github.com/awslabs/aws-sdk-go/service/autoscaling"
	"github.com/awslabs/aws-sdk-go/service/ec2"
	"github.com/awslabs/aws-sdk-go/service/elb"
	ct "github.com/daviddengcn/go-colortext"
)

var asgClient *autoscaling.AutoScaling = nil
var elbClient *elb.ELB = nil
var ec2Client *ec2.EC2 = nil

func init() {
	asgClient = autoscaling.New(nil)
	elbClient = elb.New(nil)
	ec2Client = ec2.New(nil)
}

func greenln(text string) {
	ct.ChangeColor(ct.Green, true, ct.None, false)
	fmt.Println(text)
	ct.ChangeColor(ct.White, false, ct.None, false)
}

func abort(a ...interface{}) {
	ct.ChangeColor(ct.White, true, ct.Red, true)
	fmt.Print(a...)
	ct.ChangeColor(ct.White, false, ct.None, false)
	fmt.Println("")
	os.Exit(1)
}

func abortf(text string, a ...interface{}) {
	ct.ChangeColor(ct.White, true, ct.Red, true)
	fmt.Printf(text, a...)
	ct.ChangeColor(ct.White, false, ct.None, false)
	fmt.Println("")
	os.Exit(1)
}

func getELBName(asg *amazonasg.AutoScalingGroup) string {
	numLBs := len(asg.LoadBalancerNames())
	if numLBs != 1 {
		abortf("This tool requires an auto-scaling group with exactly one load balancer, found %d", numLBs)
	}

	return asg.LoadBalancerNames()[0]
}

func getConnectionDrainingTimeout(loadBalancerName string) int64 {
	lbResp, err := elbClient.DescribeLoadBalancerAttributes(&elb.DescribeLoadBalancerAttributesInput{
		LoadBalancerName: &loadBalancerName,
	})
	if err != nil {
		abort("Error while looking up ELB Attributes:", err)
	}
	lbAttrs := lbResp.LoadBalancerAttributes
	if *lbAttrs.ConnectionDraining.Enabled != true {
		abort("Load balancer does not have connection draining enabled, will not proceed!")
	}
	return *lbAttrs.ConnectionDraining.Timeout
}

// Generate a closure that will get the Health state of an instance in an ASG.
func instanceHealthStateFunc(asgName string, instanceID string) func() string {
	return func() string {
		output, err := asgClient.DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{
			AutoScalingGroupNames: []*string{&asgName},
		})
		if err != nil {
			return "describe-error"
		}
		if len(output.AutoScalingGroups) != 1 {
			return "unknown-asg"
		}
		for _, inst := range output.AutoScalingGroups[0].Instances {
			if *inst.InstanceID == instanceID {
				return *inst.HealthStatus
			}
		}
		return "unknown-instance"
	}
}

func getInstanceStatus(instanceID string) (string, error) {
	output, err := ec2Client.DescribeInstanceStatus(&ec2.DescribeInstanceStatusInput{
		InstanceIDs: []*string{&instanceID},
	})
	if err != nil {
		return "", err
	}
	for _, inst := range output.InstanceStatuses {
		if *inst.InstanceID == instanceID {
			return *inst.InstanceState.Name, nil
		}
	}
	return "unknown-instance", nil
}

func instanceStatusFunc(instanceID string) func() string {
	return func() string {
		status, err := getInstanceStatus(instanceID)
		if err != nil {
			return "describe-error"
		}
		return status
	}
}

func removeFromELB(lbName string, instanceID string) error {
	_, err := elbClient.DeregisterInstancesFromLoadBalancer(&elb.DeregisterInstancesFromLoadBalancerInput{
		Instances: []*elb.Instance{
			&elb.Instance{InstanceID: &instanceID},
		},
		LoadBalancerName: &lbName,
	})
	return err
}

func detachFromASG(asgName string, instanceID string) error {
	d := false
	_, err := asgClient.DetachInstances(&autoscaling.DetachInstancesInput{
		AutoScalingGroupName:           &asgName,
		InstanceIDs:                    []*string{&instanceID},
		ShouldDecrementDesiredCapacity: &d,
	})
	return err
}

func terminateInstance(instanceID string) error {
	_, err := ec2Client.TerminateInstances(&ec2.TerminateInstancesInput{
		InstanceIDs: []*string{&instanceID},
	})

	return err
}

func getELBInstanceHealth(lbName string, instanceID string) string {
	output, err := elbClient.DescribeInstanceHealth(&elb.DescribeInstanceHealthInput{
		Instances: []*elb.Instance{
			&elb.Instance{InstanceID: &instanceID},
		},
		LoadBalancerName: &lbName,
	})
	if err != nil {
		return "describe-error"
	}
	if len(output.InstanceStates) != 1 {
		return "unknown-instance"
	}
	return *output.InstanceStates[0].State
}

func main() {
	greenln("This is how I roll v0.1.0.")

	config := config.ParseConfig()

	asg, err := amazonasg.Describe(config.ASGName)
	if err != nil {
		abort(err)
	}
	fmt.Println("Target LaunchConfig:", asg.LaunchConfigurationName())

	lbName := getELBName(asg)
	fmt.Println("Found Elastic Load Balancer:", lbName)

	connectionDrainingTimeout := getConnectionDrainingTimeout(lbName)
	fmt.Printf("Connection draining timeout is %d seconds.\n", connectionDrainingTimeout)

	if asg.HealthCheckType() != "ELB" {
		abort("This tool only supports ASGs that are using ELB Health Checks.")
	}

	if asg.DesiredCapacity() <= asg.MinSize() {
		abort("This tool requires that the desired capacity of the ASG be greater than it's minimum size, in order to remove outdated instances.")
	}

	// Get all instance IDs. Identify instances that are not on latest LaunchConfiguration AMI ID.
	var targetInstances []amazonasg.Instance
	for _, instance := range asg.Instances() {
		if instance.LaunchConfigurationName() != asg.LaunchConfigurationName() {
			targetInstances = append(targetInstances, instance)
		}

		if instance.HealthStatus() != "Healthy" {
			// TODO: Allow -force override?
			abortf("Instance %s is in non-Healthy state, aborting!", instance.InstanceID())
		}

		if instance.LifecycleState() != "InService" {
			abortf("Instance %s is not InService, aborting!", instance.InstanceID())
		}
	}

	if len(targetInstances) == 0 {
		abortf("All %d instances are on the correct LaunchConfig!", len(asg.Instances()))
	}

	var instanceIds []string
	for _, inst := range targetInstances {
		instanceIds = append(instanceIds, inst.InstanceID())
	}
	fmt.Printf("%d total instances, %d instance(s) to be replaced: %s\n", len(asg.Instances()), len(targetInstances), strings.Join(instanceIds, ", "))

	if config.DryRun {
		greenln("Dry run complete.")
		os.Exit(0)
	}

	// TODO: Optionally suspend autoscaling processes.

	// Track instances we know exist in the ASG.
	var knownInstanceIDs []string = make([]string, len(asg.InstanceIDs()))
	copy(knownInstanceIDs, asg.InstanceIDs())

	for _, inst := range targetInstances {
		fmt.Printf("Detaching %s from the Auto-Scaling Group.\n", inst.InstanceID())
		detachFromASG(config.ASGName, inst.InstanceID())

		waitfor.MissingString(
			fmt.Sprintf("Waiting for %s to be removed from the Auto-Scaling Group.", inst.InstanceID()),
			func() []string {
				if asg, err := amazonasg.Describe(config.ASGName); err != nil {
					// On error, return an array containing the instance we're waiting
					// for removal on, so that we retry.
					return []string{inst.InstanceID()}
				} else {
					return asg.InstanceIDs()
				}
			},
			inst.InstanceID())

		newInstanceID := waitfor.AdditionalString(
			"Waiting for a new instance to appear in the ASG.",
			func() []string {
				if asg, err := amazonasg.Describe(config.ASGName); err != nil {
					return []string{}
				} else {
					return asg.InstanceIDs()
				}
			},
			knownInstanceIDs)

		waitfor.Strings(
			fmt.Sprintf("Waiting for the the new instance (%s) to be InService on the Load Balancer.", newInstanceID),
			func() string {
				return getELBInstanceHealth(lbName, newInstanceID)
			},
			[]string{"InService"})

		fmt.Printf("(----) Terminating detached instance %s.\n", inst.InstanceID())
		terminateInstance(inst.InstanceID())

		// Add the new instance, so it doesn't appear to be new on the next iteration.
		knownInstanceIDs = append(knownInstanceIDs, newInstanceID)
	}

	greenln("All instances are now using the latest LaunchConfig!")
}
