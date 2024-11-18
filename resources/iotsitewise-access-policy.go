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

const IoTSiteWiseAccessPolicyResource = "IoTSiteWiseAccessPolicy"

func init() {
	registry.Register(&registry.Registration{
		Name:   IoTSiteWiseAccessPolicyResource,
		Scope:  nuke.Account,
		Lister: &IoTSiteWiseAccessPolicyLister{},
	})
}

type IoTSiteWiseAccessPolicyLister struct{}

func (l *IoTSiteWiseAccessPolicyLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) { //nolint:gocyclo
	opts := o.(*nuke.ListerOpts)

	svc := iotsitewise.New(opts.Session)
	resources := make([]resource.Resource, 0)

	// Policies can be attached either to portal or projects
	// List portal and portal policies
	listPortalsParams := &iotsitewise.ListPortalsInput{
		MaxResults: aws.Int64(25),
	}
	for {
		listPortalsResp, err := svc.ListPortals(listPortalsParams)
		if err != nil {
			return nil, err
		}
		for _, portalItem := range listPortalsResp.PortalSummaries {
			// Got portals
			listProjectsParams := &iotsitewise.ListProjectsInput{
				PortalId:   portalItem.Id,
				MaxResults: aws.Int64(25),
			}

			// List portal policies
			listPortalPoliciesParam := &iotsitewise.ListAccessPoliciesInput{
				ResourceId:   portalItem.Id,
				ResourceType: &([]string{string(iotsitewise.ResourceTypePortal)}[0]),
				MaxResults:   aws.Int64(25),
			}

			for {
				listPortalPoliciesResp, err := svc.ListAccessPolicies(listPortalPoliciesParam)
				if err != nil {
					return nil, err
				}
				for _, item := range listPortalPoliciesResp.AccessPolicySummaries {
					resources = append(resources, &IoTSiteWiseAccessPolicy{
						svc: svc,
						ID:  item.Id,
					})
				}

				if listPortalPoliciesResp.NextToken == nil {
					break
				}

				listPortalPoliciesParam.NextToken = listPortalPoliciesResp.NextToken
			}

			// will also search inside projects
			for {
				listProjectsResp, err := svc.ListProjects(listProjectsParams)
				if err != nil {
					return nil, err
				}
				for _, projectItem := range listProjectsResp.ProjectSummaries {
					// List project policies
					listProjectPoliciesParams := &iotsitewise.ListAccessPoliciesInput{
						ResourceId:   projectItem.Id,
						ResourceType: &([]string{string(iotsitewise.ResourceTypeProject)}[0]),
						MaxResults:   aws.Int64(25),
					}

					for {
						listProjectPoliciesResp, err := svc.ListAccessPolicies(listProjectPoliciesParams)
						if err != nil {
							return nil, err
						}
						for _, item := range listProjectPoliciesResp.AccessPolicySummaries {
							resources = append(resources, &IoTSiteWiseAccessPolicy{
								svc: svc,
								ID:  item.Id,
							})
						}

						if listProjectPoliciesResp.NextToken == nil {
							break
						}

						listProjectPoliciesParams.NextToken = listProjectPoliciesResp.NextToken
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

type IoTSiteWiseAccessPolicy struct {
	svc *iotsitewise.IoTSiteWise
	ID  *string
}

func (r *IoTSiteWiseAccessPolicy) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *IoTSiteWiseAccessPolicy) Remove(_ context.Context) error {
	_, err := r.svc.DeleteAccessPolicy(&iotsitewise.DeleteAccessPolicyInput{
		AccessPolicyId: r.ID,
	})

	return err
}

func (r *IoTSiteWiseAccessPolicy) String() string {
	return *r.ID
}
