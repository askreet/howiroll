package amazonasg

import (
	"errors"

	"github.com/awslabs/aws-sdk-go/service/autoscaling"
)

var asgClient *autoscaling.AutoScaling

func init() {
	asgClient = autoscaling.New(nil)
}

type AutoScalingGroup struct {
	internal *autoscaling.AutoScalingGroup
}

func Describe(asgName string) (*AutoScalingGroup, error) {
	output, err := asgClient.DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []*string{&asgName},
	})

	if err != nil {
		return nil, err
	}

	if len(output.AutoScalingGroups) != 1 {
		return nil, errors.New("No such AutoScaling Group: " + asgName)
	}

	return &AutoScalingGroup{
		internal: output.AutoScalingGroups[0],
	}, nil
}

func (asg *AutoScalingGroup) HasLaunchConfiguration() bool {
	return asg.internal.LaunchConfigurationName != nil
}

func (asg *AutoScalingGroup) LaunchConfigurationName() string {
	if asg.HasLaunchConfiguration() {
		return *asg.internal.LaunchConfigurationName
	} else {
		return ""
	}
}

func (asg *AutoScalingGroup) InstanceIDs() []string {
	var result []string
	for _, inst := range asg.internal.Instances {
		result = append(result, *inst.InstanceID)
	}
	return result
}

func (asg *AutoScalingGroup) LoadBalancerNames() []string {
	var result []string
	for _, lbName := range asg.internal.LoadBalancerNames {
		result = append(result, *lbName)
	}
	return result
}

func (asg *AutoScalingGroup) Instances() []Instance {
	var result []Instance
	for _, inst := range asg.internal.Instances {
		result = append(result, Instance{internal: *inst})
	}
	return result
}

func (asg *AutoScalingGroup) HealthCheckType() string {
	if asg.internal.HealthCheckType == nil {
		return ""
	} else {
		return *asg.internal.HealthCheckType
	}
}

func (asg *AutoScalingGroup) DesiredCapacity() int64 {
	if asg.internal.DesiredCapacity == nil {
		return 0
	} else {
		return *asg.internal.DesiredCapacity
	}
}

func (asg *AutoScalingGroup) MinSize() int64 {
	if asg.internal.MinSize == nil {
		return 0
	} else {
		return *asg.internal.MinSize
	}
}

type Instance struct {
	internal autoscaling.Instance
}

func (inst *Instance) HasLaunchConfiguration() bool {
	return inst.internal.LaunchConfigurationName != nil
}

func (inst *Instance) LaunchConfigurationName() string {
	if inst.HasLaunchConfiguration() {
		return *inst.internal.LaunchConfigurationName
	}
	return ""
}

func (inst *Instance) HealthStatus() string {
	if inst.internal.HealthStatus == nil {
		return ""
	} else {
		return *inst.internal.HealthStatus
	}
}

func (inst *Instance) InstanceID() string {
	return *inst.internal.InstanceID
}

func (inst *Instance) LifecycleState() string {
	if inst.internal.LifecycleState == nil {
		return ""
	} else {
		return *inst.internal.LifecycleState
	}
}
