package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iottwinmaker"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const IoTTwinMakerEntityResource = "IoTTwinMakerEntity"

func init() {
	registry.Register(&registry.Registration{
		Name:     IoTTwinMakerEntityResource,
		Scope:    nuke.Account,
		Resource: &IoTTwinMakerEntity{},
		Lister:   &IoTTwinMakerEntityLister{},
	})
}

type IoTTwinMakerEntityLister struct{}

func (l *IoTTwinMakerEntityLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := iottwinmaker.New(opts.Session)
	resources := make([]resource.Resource, 0)

	// Require to have workspaces identifiers to query entities
	workspaceListResponse, err := ListWorkspacesEntities(svc)

	if err != nil {
		return nil, err
	}

	for _, workspaceResponse := range workspaceListResponse {
		params := &iottwinmaker.ListEntitiesInput{
			WorkspaceId: workspaceResponse.WorkspaceId,
			MaxResults:  aws.Int64(25),
		}

		for {
			resp, err := svc.ListEntities(params)
			if err != nil {
				return nil, err
			}

			for _, item := range resp.EntitySummaries {
				// We must filter out amazon-owned component types when querying tags,
				// because their ARN format causes ListTagsForResource to vail with validation error
				resources = append(resources, &IoTTwinMakerEntity{
					svc:         svc,
					ID:          item.EntityId,
					Name:        item.EntityName,
					Status:      item.Status.State,
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
func ListWorkspacesEntities(svc *iottwinmaker.IoTTwinMaker) ([]*iottwinmaker.WorkspaceSummary, error) {
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

type IoTTwinMakerEntity struct {
	svc         *iottwinmaker.IoTTwinMaker
	ID          *string
	Name        *string
	Status      *string
	WorkspaceID *string
}

func (r *IoTTwinMakerEntity) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *IoTTwinMakerEntity) Remove(_ context.Context) error {
	_, err := r.svc.DeleteEntity(&iottwinmaker.DeleteEntityInput{
		EntityId:    r.ID,
		WorkspaceId: r.WorkspaceID,
		IsRecursive: aws.Bool(true),
	})

	return err
}

func (r *IoTTwinMakerEntity) String() string {
	return *r.ID
}
