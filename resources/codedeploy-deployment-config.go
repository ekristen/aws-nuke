package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/service/codedeploy"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const CodeDeployDeploymentConfigResource = "CodeDeployDeploymentConfig"

func init() {
	registry.Register(&registry.Registration{
		Name:     CodeDeployDeploymentConfigResource,
		Scope:    nuke.Account,
		Resource: &CodeDeployDeploymentConfig{},
		Lister:   &CodeDeployDeploymentConfigLister{},
	})
}

type CodeDeployDeploymentConfigLister struct{}

func (l *CodeDeployDeploymentConfigLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := codedeploy.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &codedeploy.ListDeploymentConfigsInput{}

	for {
		resp, err := svc.ListDeploymentConfigs(params)
		if err != nil {
			return nil, err
		}

		for _, config := range resp.DeploymentConfigsList {
			resources = append(resources, &CodeDeployDeploymentConfig{
				svc:  svc,
				Name: config,
			})
		}

		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return resources, nil
}

type CodeDeployDeploymentConfig struct {
	svc  *codedeploy.CodeDeploy
	Name *string
}

func (r *CodeDeployDeploymentConfig) Filter() error {
	if strings.HasPrefix(*r.Name, "CodeDeployDefault") {
		return fmt.Errorf("cannot delete default codedeploy config")
	}
	return nil
}

func (r *CodeDeployDeploymentConfig) Remove(_ context.Context) error {
	_, err := r.svc.DeleteDeploymentConfig(&codedeploy.DeleteDeploymentConfigInput{
		DeploymentConfigName: r.Name,
	})

	return err
}

func (r *CodeDeployDeploymentConfig) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *CodeDeployDeploymentConfig) String() string {
	return *r.Name
}
