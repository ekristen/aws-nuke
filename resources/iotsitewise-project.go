package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iotsitewise"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const IoTSiteWiseProjectResource = "IoTSiteWiseProject"

func init() {
	registry.Register(&registry.Registration{
		Name:   IoTSiteWiseProjectResource,
		Scope:  nuke.Account,
		Lister: &IoTSiteWiseProjectLister{},
		DependsOn: []string{
			IoTSiteWiseDashboardResource,
			IoTSiteWiseAccessPolicyResource,
		},
	})
}

type IoTSiteWiseProjectLister struct{}

func (l *IoTSiteWiseProjectLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := iotsitewise.New(opts.Session)
	resources := make([]resource.Resource, 0)

	// To get projects, we must list all portals
	listPortalsParams := &iotsitewise.ListPortalsInput{
		MaxResults: aws.Int64(25),
	}
	for {
		listPortalsResp, err := svc.ListPortals(listPortalsParams)
		if err != nil {
			return nil, err
		}
		for _, portalItem := range listPortalsResp.PortalSummaries {
			// Got portals, will search for projects
			listProjectsParams := &iotsitewise.ListProjectsInput{
				PortalId:   portalItem.Id,
				MaxResults: aws.Int64(25),
			}

			for {
				listProjectsResp, err := svc.ListProjects(listProjectsParams)
				if err != nil {
					return nil, err
				}
				for _, projectItem := range listProjectsResp.ProjectSummaries {
					resources = append(resources, &IoTSiteWiseProject{
						svc:  svc,
						ID:   projectItem.Id,
						Name: projectItem.Name,
					})
				}

				if listProjectsResp.NextToken == nil {
					break
				}

				listProjectsParams.NextToken = listProjectsResp.NextToken
			}
		}

		if listPortalsResp.NextToken == nil {
			break
		}

		listPortalsParams.NextToken = listPortalsResp.NextToken
	}

	return resources, nil
}

type IoTSiteWiseProject struct {
	svc  *iotsitewise.IoTSiteWise
	ID   *string
	Name *string
}

func (r *IoTSiteWiseProject) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *IoTSiteWiseProject) Remove(_ context.Context) error {
	_, err := r.svc.DeleteProject(&iotsitewise.DeleteProjectInput{
		ProjectId: r.ID,
	})

	return err
}

func (r *IoTSiteWiseProject) String() string {
	return *r.ID
}
