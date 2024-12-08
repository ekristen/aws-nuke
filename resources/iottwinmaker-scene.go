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

const IoTTwinMakerSceneResource = "IoTTwinMakerScene"

func init() {
	registry.Register(&registry.Registration{
		Name:     IoTTwinMakerSceneResource,
		Scope:    nuke.Account,
		Resource: &IoTTwinMakerScene{},
		Lister:   &IoTTwinMakerSceneLister{},
	})
}

type IoTTwinMakerSceneLister struct {
	IoTTwinMaker
}

func (l *IoTTwinMakerSceneLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	resources := make([]resource.Resource, 0)

	if !l.IsSupportedRegion(opts.Region.Name) {
		return resources, nil
	}

	svc := iottwinmaker.New(opts.Session)

	// Require to have workspaces identifiers to query scenes
	workspaceListResponse, err := ListWorkspacesScene(svc)

	if err != nil {
		return nil, err
	}

	for _, workspaceResponse := range workspaceListResponse {
		params := &iottwinmaker.ListScenesInput{
			WorkspaceId: workspaceResponse.WorkspaceId,
			MaxResults:  aws.Int64(25),
		}

		for {
			resp, err := svc.ListScenes(params)
			if err != nil {
				return nil, err
			}

			for _, item := range resp.SceneSummaries {
				resources = append(resources, &IoTTwinMakerScene{
					svc:         svc,
					ID:          item.SceneId,
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
func ListWorkspacesScene(svc *iottwinmaker.IoTTwinMaker) ([]*iottwinmaker.WorkspaceSummary, error) {
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

type IoTTwinMakerScene struct {
	svc         *iottwinmaker.IoTTwinMaker
	ID          *string
	WorkspaceID *string
}

func (r *IoTTwinMakerScene) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *IoTTwinMakerScene) Remove(_ context.Context) error {
	_, err := r.svc.DeleteScene(&iottwinmaker.DeleteSceneInput{
		SceneId:     r.ID,
		WorkspaceId: r.WorkspaceID,
	})

	return err
}

func (r *IoTTwinMakerScene) String() string {
	return *r.ID
}
