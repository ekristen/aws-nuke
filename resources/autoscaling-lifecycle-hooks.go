package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/autoscaling"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

// TODO: review if this should be prefixed with Autoscaling

const LifecycleHookResource = "LifecycleHook"

func init() {
	registry.Register(&registry.Registration{
		Name:   LifecycleHookResource,
		Scope:  nuke.Account,
		Lister: &LifecycleHookLister{},
	})
}

type LifecycleHookLister struct{}

func (l *LifecycleHookLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := autoscaling.New(opts.Session)

	asgResp, err := svc.DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{})
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, 0)
	for _, asg := range asgResp.AutoScalingGroups {
		lchResp, err := svc.DescribeLifecycleHooks(&autoscaling.DescribeLifecycleHooksInput{
			AutoScalingGroupName: asg.AutoScalingGroupName,
		})
		if err != nil {
			return nil, err
		}

		for _, lch := range lchResp.LifecycleHooks {
			resources = append(resources, &LifecycleHook{
				svc:                  svc,
				lifecycleHookName:    lch.LifecycleHookName,
				autoScalingGroupName: lch.AutoScalingGroupName,
			})
		}
	}

	return resources, nil
}

type LifecycleHook struct {
	svc                  *autoscaling.AutoScaling
	lifecycleHookName    *string
	autoScalingGroupName *string
}

func (lch *LifecycleHook) Remove(_ context.Context) error {
	params := &autoscaling.DeleteLifecycleHookInput{
		AutoScalingGroupName: lch.autoScalingGroupName,
		LifecycleHookName:    lch.lifecycleHookName,
	}

	_, err := lch.svc.DeleteLifecycleHook(params)
	if err != nil {
		return err
	}

	return nil
}

func (lch *LifecycleHook) String() string {
	return *lch.lifecycleHookName
}
