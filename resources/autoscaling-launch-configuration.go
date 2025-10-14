package resources

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/autoscaling" //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/autoscaling/autoscalingiface"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const AutoScalingLaunchConfigurationResource = "AutoScalingLaunchConfiguration"

func init() {
	registry.Register(&registry.Registration{
		Name:     AutoScalingLaunchConfigurationResource,
		Scope:    nuke.Account,
		Resource: &AutoScalingLaunchConfiguration{},
		Lister:   &AutoScalingLaunchConfigurationLister{},
		DeprecatedAliases: []string{
			"LaunchConfiguration",
		},
	})
}

type AutoScalingLaunchConfigurationLister struct {
	mockSvc autoscalingiface.AutoScalingAPI
}

func (l *AutoScalingLaunchConfigurationLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	resources := make([]resource.Resource, 0)

	var svc autoscalingiface.AutoScalingAPI
	if l.mockSvc != nil {
		svc = l.mockSvc
	} else {
		svc = autoscaling.New(opts.Session)
	}

	params := &autoscaling.DescribeLaunchConfigurationsInput{}
	err := svc.DescribeLaunchConfigurationsPages(params,
		func(page *autoscaling.DescribeLaunchConfigurationsOutput, lastPage bool) bool {
			for _, launchConfig := range page.LaunchConfigurations {
				resources = append(resources, &AutoScalingLaunchConfiguration{
					svc:         svc,
					Name:        launchConfig.LaunchConfigurationName,
					CreatedTime: launchConfig.CreatedTime,
				})
			}
			return !lastPage
		})

	if err != nil {
		return nil, err
	}

	return resources, nil
}

type AutoScalingLaunchConfiguration struct {
	svc         autoscalingiface.AutoScalingAPI
	Name        *string
	CreatedTime *time.Time
}

func (r *AutoScalingLaunchConfiguration) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *AutoScalingLaunchConfiguration) Remove(_ context.Context) error {
	params := &autoscaling.DeleteLaunchConfigurationInput{
		LaunchConfigurationName: r.Name,
	}

	_, err := r.svc.DeleteLaunchConfiguration(params)
	if err != nil {
		return err
	}

	return nil
}

func (r *AutoScalingLaunchConfiguration) String() string {
	return *r.Name
}
