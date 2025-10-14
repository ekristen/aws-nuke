package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"                  //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/iottwinmaker" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const IoTTwinMakerWorkspaceResource = "IoTTwinMakerWorkspace"

func init() {
	registry.Register(&registry.Registration{
		Name:     IoTTwinMakerWorkspaceResource,
		Scope:    nuke.Account,
		Resource: &IoTTwinMakerWorkspace{},
		Lister:   &IoTTwinMakerWorkspaceLister{},
		DependsOn: []string{
			IoTTwinMakerComponentTypeResource,
			IoTTwinMakerEntityResource,
			IoTTwinMakerSceneResource,
			IoTTwinMakerSyncJobResource,
		},
	})
}

type IoTTwinMakerWorkspaceLister struct {
	IoTTwinMaker
}

func (l *IoTTwinMakerWorkspaceLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	resources := make([]resource.Resource, 0)

	if !l.IsSupportedRegion(opts.Region.Name) {
		return resources, nil
	}

	svc := iottwinmaker.New(opts.Session)

	params := &iottwinmaker.ListWorkspacesInput{
		MaxResults: aws.Int64(25),
	}

	for {
		resp, err := svc.ListWorkspaces(params)
		if err != nil {
			return nil, err
		}

		for _, item := range resp.WorkspaceSummaries {
			tagResp, err := svc.ListTagsForResource(
				&iottwinmaker.ListTagsForResourceInput{
					ResourceARN: item.Arn,
				})
			if err != nil {
				return nil, err
			}

			resources = append(resources, &IoTTwinMakerWorkspace{
				svc:  svc,
				ID:   item.WorkspaceId,
				Tags: tagResp.Tags,
			})
		}

		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return resources, nil
}

type IoTTwinMakerWorkspace struct {
	svc  *iottwinmaker.IoTTwinMaker
	ID   *string
	Tags map[string]*string
}

func (r *IoTTwinMakerWorkspace) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *IoTTwinMakerWorkspace) Remove(_ context.Context) error {
	_, err := r.svc.DeleteWorkspace(&iottwinmaker.DeleteWorkspaceInput{
		WorkspaceId: r.ID,
	})

	return err
}

func (r *IoTTwinMakerWorkspace) String() string {
	return *r.ID
}
