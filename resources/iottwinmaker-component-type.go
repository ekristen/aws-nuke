package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iottwinmaker"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const IoTTwinMakerComponentTypeResource = "IoTTwinMakerComponentType"

func init() {
	registry.Register(&registry.Registration{
		Name:   IoTTwinMakerComponentTypeResource,
		Scope:  nuke.Account,
		Lister: &IoTTwinMakerComponentTypeLister{},
	})
}

type IoTTwinMakerComponentTypeLister struct{}

func (l *IoTTwinMakerComponentTypeLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := iottwinmaker.New(opts.Session)
	resources := make([]resource.Resource, 0)

	// Require to have workspaces identifiers to query components
	workspaceListResponse, err := ListWorkspacesComponentType(svc)

	if err != nil {
		return nil, err
	}

	for _, workspaceResponse := range workspaceListResponse {
		params := &iottwinmaker.ListComponentTypesInput{
			WorkspaceId: workspaceResponse.WorkspaceId,
			MaxResults:  aws.Int64(25),
		}

		for {
			resp, err := svc.ListComponentTypes(params)
			if err != nil {
				return nil, err
			}

			for _, item := range resp.ComponentTypeSummaries {
				// We must filter out amazon-owned component types when querying tags,
				// because their ARN format causes ListTagsForResource to fail with validation error
				tags := make(map[string]*string)
				if !strings.Contains(*item.Arn, "AmazonOwnedTypesWorkspace") {
					tagResp, err := svc.ListTagsForResource(
						&iottwinmaker.ListTagsForResourceInput{
							ResourceARN: item.Arn,
						})
					if err != nil {
						return nil, err
					}
					tags = tagResp.Tags
				}

				resources = append(resources, &IoTTwinMakerComponentType{
					svc:         svc,
					ID:          item.ComponentTypeId,
					arn:         item.Arn,
					Tags:        tags,
					WorkspaceID: workspaceResponse.WorkspaceId,
				})
			}

			if resp.NextToken == nil {
				break
			}

			params.NextToken = resp.NextToken
		}
	}

	return resources, nil
}

// Utility function to list workspaces
func ListWorkspacesComponentType(svc *iottwinmaker.IoTTwinMaker) ([]*iottwinmaker.WorkspaceSummary, error) {
	resources := make([]*iottwinmaker.WorkspaceSummary, 0)
	params := &iottwinmaker.ListWorkspacesInput{
		MaxResults: aws.Int64(25),
	}
	for {
		resp, err := svc.ListWorkspaces(params)
		if err != nil {
			return nil, err
		}
		resources = append(resources, resp.WorkspaceSummaries...)
		if resp.NextToken == nil {
			break
		}
		params.NextToken = resp.NextToken
	}
	return resources, nil
}

type IoTTwinMakerComponentType struct {
	svc         *iottwinmaker.IoTTwinMaker
	ID          *string
	Tags        map[string]*string
	WorkspaceID *string
	arn         *string
}

func (r *IoTTwinMakerComponentType) Filter() error {
	if strings.Contains(*r.arn, "AmazonOwnedTypesWorkspace") {
		return fmt.Errorf("cannot delete pre-defined component type")
	}
	return nil
}

func (r *IoTTwinMakerComponentType) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *IoTTwinMakerComponentType) Remove(_ context.Context) error {
	_, err := r.svc.DeleteComponentType(&iottwinmaker.DeleteComponentTypeInput{
		ComponentTypeId: r.ID,
		WorkspaceId:     r.WorkspaceID,
	})

	return err
}

func (r *IoTTwinMakerComponentType) String() string {
	return *r.ID
}
