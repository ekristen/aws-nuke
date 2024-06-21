package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appconfig"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

type AppConfigDeploymentStrategy struct {
	svc  *appconfig.AppConfig
	id   *string
	name *string
}

const AppConfigDeploymentStrategyResource = "AppConfigDeploymentStrategy"

func init() {
	registry.Register(&registry.Registration{
		Name:   AppConfigDeploymentStrategyResource,
		Scope:  nuke.Account,
		Lister: &AppConfigDeploymentStrategyLister{},
	})
}

type AppConfigDeploymentStrategyLister struct{}

func (l *AppConfigDeploymentStrategyLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := appconfig.New(opts.Session)
	resources := make([]resource.Resource, 0)
	params := &appconfig.ListDeploymentStrategiesInput{
		MaxResults: aws.Int64(50),
	}
	err := svc.ListDeploymentStrategiesPages(params, func(page *appconfig.ListDeploymentStrategiesOutput, lastPage bool) bool {
		for _, item := range page.Items {
			resources = append(resources, &AppConfigDeploymentStrategy{
				svc:  svc,
				id:   item.Id,
				name: item.Name,
			})
		}
		return true
	})
	if err != nil {
		return nil, err
	}
	return resources, nil
}

func (f *AppConfigDeploymentStrategy) Filter() error {
	if strings.HasPrefix(*f.name, "AppConfig.") {
		return fmt.Errorf("cannot delete predefined Deployment Strategy")
	}
	return nil
}

func (f *AppConfigDeploymentStrategy) Remove(_ context.Context) error {
	_, err := f.svc.DeleteDeploymentStrategy(&appconfig.DeleteDeploymentStrategyInput{
		DeploymentStrategyId: f.id,
	})
	return err
}

func (f *AppConfigDeploymentStrategy) Properties() types.Properties {
	return types.NewProperties().
		Set("ID", f.id).
		Set("Name", f.name)
}
