package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/autoscaling/autoscalingiface"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const AutoScalingLifecycleHookResource = "AutoScalingLifecycleHook"

func init() {
	registry.Register(&registry.Registration{
		Name:   AutoScalingLifecycleHookResource,
		Scope:  nuke.Account,
		Lister: &AutoScalingLifecycleHookLister{},
		DeprecatedAliases: []string{
			"LifecycleHook",
		},
	})
}

type AutoScalingLifecycleHookLister struct {
	mockSvc autoscalingiface.AutoScalingAPI
}

func (l *AutoScalingLifecycleHookLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	var svc autoscalingiface.AutoScalingAPI
	if l.mockSvc != nil {
		svc = l.mockSvc
	} else {
		svc = autoscaling.New(opts.Session)
	}

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
			resources = append(resources, &AutoScalingLifecycleHook{
				svc:       svc,
				Name:      lch.LifecycleHookName,
				GroupName: lch.AutoScalingGroupName,
			})
		}
	}

	return resources, nil
}

type AutoScalingLifecycleHook struct {
	svc       autoscalingiface.AutoScalingAPI
	Name      *string
	GroupName *string
}

func (r *AutoScalingLifecycleHook) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *AutoScalingLifecycleHook) Remove(_ context.Context) error {
	params := &autoscaling.DeleteLifecycleHookInput{
		AutoScalingGroupName: r.GroupName,
		LifecycleHookName:    r.Name,
	}

	_, err := r.svc.DeleteLifecycleHook(params)
	if err != nil {
		return err
	}

	return nil
}

func (r *AutoScalingLifecycleHook) String() string {
	return *r.Name
}
