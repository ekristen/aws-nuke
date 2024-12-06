package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/codedeploy"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const CodeDeployDeploymentGroupResource = "CodeDeployDeploymentGroup"

func init() {
	registry.Register(&registry.Registration{
		Name:     CodeDeployDeploymentGroupResource,
		Scope:    nuke.Account,
		Resource: &CodeDeployDeploymentGroup{},
		Lister:   &CodeDeployDeploymentGroupLister{},
	})
}

type CodeDeployDeploymentGroupLister struct{}

func (l *CodeDeployDeploymentGroupLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	var resources []resource.Resource

	svc := codedeploy.New(opts.Session)

	params := &codedeploy.ListApplicationsInput{}

	for {
		appResp, err := svc.ListApplications(params)
		if err != nil {
			return nil, err
		}

		for _, appName := range appResp.Applications {
			// For each application, list deployment groups
			deploymentGroupParams := &codedeploy.ListDeploymentGroupsInput{
				ApplicationName: appName,
			}
			deploymentGroupResp, err := svc.ListDeploymentGroups(deploymentGroupParams)
			if err != nil {
				return nil, err
			}

			for _, group := range deploymentGroupResp.DeploymentGroups {
				resources = append(resources, &CodeDeployDeploymentGroup{
					svc:             svc,
					Name:            group,
					ApplicationName: appName,
				})
			}
		}

		if appResp.NextToken == nil {
			break
		}

		params.NextToken = appResp.NextToken
	}

	return resources, nil
}

type CodeDeployDeploymentGroup struct {
	svc             *codedeploy.CodeDeploy
	Name            *string
	ApplicationName *string
}

func (r *CodeDeployDeploymentGroup) Remove(_ context.Context) error {
	_, err := r.svc.DeleteDeploymentGroup(&codedeploy.DeleteDeploymentGroupInput{
		ApplicationName:     r.ApplicationName,
		DeploymentGroupName: r.Name,
	})

	return err
}

func (r *CodeDeployDeploymentGroup) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *CodeDeployDeploymentGroup) String() string {
	return *r.Name
}
