package resources

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/amp"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const AMPWorkspaceResource = "AMPWorkspace"

func init() {
	registry.Register(&registry.Registration{
		Name:     AMPWorkspaceResource,
		Scope:    nuke.Account,
		Resource: &AMPWorkspace{},
		Lister:   &AMPWorkspaceLister{},
	})
}

type AMPWorkspaceLister struct{}

func (l *AMPWorkspaceLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := amp.NewFromConfig(*opts.Config)
	resources := make([]resource.Resource, 0)

	paginator := amp.NewListWorkspacesPaginator(svc, &amp.ListWorkspacesInput{})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, ws := range page.Workspaces {
			resources = append(resources, &AMPWorkspace{
				svc:            svc,
				WorkspaceAlias: ws.Alias,
				WorkspaceARN:   ws.Arn,
				WorkspaceId:    ws.WorkspaceId,
			})
		}
	}

	return resources, nil
}

type AMPWorkspace struct {
	svc            *amp.Client
	WorkspaceAlias *string
	WorkspaceARN   *string
	WorkspaceId    *string
}

func (f *AMPWorkspace) Remove(ctx context.Context) error {
	_, err := f.svc.DeleteWorkspace(ctx, &amp.DeleteWorkspaceInput{
		WorkspaceId: f.WorkspaceId,
	})

	return err
}

func (f *AMPWorkspace) Properties() types.Properties {
	return types.NewPropertiesFromStruct(f)
}
