# This is how I roll...



`howiroll` is a tool for rolling out an updating LaunchConfiguration to
instances in an Amazon Web Services Auto-Scaling Group.

## Requirements

- You must have a configured Auto-Scaling Group with exactly one ELB attached.
- Connection Draining must be enabled on the ELB.
- The ASG must use ELB Health Checks.

## Installation

This software can be built with `gb`. Download `gb` at http://getgb.io. Once
you have done that, you can build the software with `gb build`.

## Usage

AWS configuration is taken directly from the environment by the AWS SDK for Go,
so you must set `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, and `AWS_REGION`.

Then, pass the Auto-Scaling Group name on the commandline:
```
howiroll -asg-name MyAutoScalingGroup
```

## TODO / Roadmap

### v0.2

- Code cleanup, comments, tests.
- Find a cool logo, possibly involving a broken down wagon cart, or similar.

### v0.4

- Support suspending Scaling Policies during rollout.
- Support timeouts.

### v1.0

- Support N-at-a-time rollouts for large Auto-Scaling Groups.

## Contributing

If you have a feature request or find a bug, please submit a GitHub Issue on the
right.

- Fork this repository.
- Create a feature branch.
- Submit a pull request against master.
