package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"                 //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/iotsitewise" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const IoTSiteWiseDashboardResource = "IoTSiteWiseDashboard"

func init() {
	registry.Register(&registry.Registration{
		Name:     IoTSiteWiseDashboardResource,
		Scope:    nuke.Account,
		Resource: &IoTSiteWiseDashboard{},
		Lister:   &IoTSiteWiseDashboardLister{},
	})
}

type IoTSiteWiseDashboardLister struct{}

func (l *IoTSiteWiseDashboardLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := iotsitewise.New(opts.Session)
	resources := make([]resource.Resource, 0)

	// Dashboards can be listed from each project
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
					// Got projects, finally get dashboards
					listDashboardParams := &iotsitewise.ListDashboardsInput{
						ProjectId:  projectItem.Id,
						MaxResults: aws.Int64(25),
					}

					for {
						listDashboardResp, err := svc.ListDashboards(listDashboardParams)
						if err != nil {
							return nil, err
						}
						for _, dashboardItem := range listDashboardResp.DashboardSummaries {
							resources = append(resources, &IoTSiteWiseDashboard{
								svc:  svc,
								ID:   dashboardItem.Id,
								Name: dashboardItem.Name,
							})
						}

						if listDashboardResp.NextToken == nil {
							break
						}

						listDashboardParams.NextToken = listDashboardResp.NextToken
					}
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

type IoTSiteWiseDashboard struct {
	svc  *iotsitewise.IoTSiteWise
	ID   *string
	Name *string
}

func (r *IoTSiteWiseDashboard) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *IoTSiteWiseDashboard) Remove(_ context.Context) error {
	_, err := r.svc.DeleteDashboard(&iotsitewise.DeleteDashboardInput{
		DashboardId: r.ID,
	})

	return err
}

func (r *IoTSiteWiseDashboard) String() string {
	return *r.ID
}
