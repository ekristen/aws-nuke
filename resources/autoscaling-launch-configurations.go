package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/autoscaling"

	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

// TODO: review if this should be prefixed with Autoscaling

const LaunchConfigurationResource = "LaunchConfiguration"

func init() {
	resource.Register(resource.Registration{
		Name:   LaunchConfigurationResource,
		Scope:  nuke.Account,
		Lister: &LaunchConfigurationLister{},
	})
}

type LaunchConfigurationLister struct{}

func (l *LaunchConfigurationLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	resources := make([]resource.Resource, 0)
	svc := autoscaling.New(opts.Session)

	params := &autoscaling.DescribeLaunchConfigurationsInput{}
	err := svc.DescribeLaunchConfigurationsPages(params,
		func(page *autoscaling.DescribeLaunchConfigurationsOutput, lastPage bool) bool {
			for _, launchConfig := range page.LaunchConfigurations {
				resources = append(resources, &LaunchConfiguration{
					svc:  svc,
					name: launchConfig.LaunchConfigurationName,
				})
			}
			return !lastPage
		})

	if err != nil {
		return nil, err
	}

	return resources, nil
}

type LaunchConfiguration struct {
	svc  *autoscaling.AutoScaling
	name *string
}

func (c *LaunchConfiguration) Remove(_ context.Context) error {
	params := &autoscaling.DeleteLaunchConfigurationInput{
		LaunchConfigurationName: c.name,
	}

	_, err := c.svc.DeleteLaunchConfiguration(params)
	if err != nil {
		return err
	}

	return nil
}

func (c *LaunchConfiguration) String() string {
	return *c.name
}
