package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/prometheusservice"

	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const AMPWorkspaceResource = "AMPWorkspace"

func init() {
	resource.Register(resource.Registration{
		Name:   AMPWorkspaceResource,
		Scope:  nuke.Account,
		Lister: &AMPWorkspaceLister{},
	})
}

type AMPWorkspaceLister struct{}

func (l *AMPWorkspaceLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := prometheusservice.New(opts.Session)
	resources := make([]resource.Resource, 0)

	var ampWorkspaces []*prometheusservice.WorkspaceSummary
	err := svc.ListWorkspacesPages(
		&prometheusservice.ListWorkspacesInput{},
		func(page *prometheusservice.ListWorkspacesOutput, lastPage bool) bool {
			ampWorkspaces = append(ampWorkspaces, page.Workspaces...)
			return true
		},
	)
	if err != nil {
		return nil, err
	}

	for _, ws := range ampWorkspaces {
		resources = append(resources, &AMPWorkspace{
			svc:            svc,
			workspaceAlias: ws.Alias,
			workspaceARN:   ws.Arn,
			workspaceId:    ws.WorkspaceId,
		})
	}

	return resources, nil
}

type AMPWorkspace struct {
	svc            *prometheusservice.PrometheusService
	workspaceAlias *string
	workspaceARN   *string
	workspaceId    *string
}

func (f *AMPWorkspace) Remove(_ context.Context) error {
	_, err := f.svc.DeleteWorkspace(&prometheusservice.DeleteWorkspaceInput{
		WorkspaceId: f.workspaceId,
	})

	return err
}

func (f *AMPWorkspace) Properties() types.Properties {
	properties := types.NewProperties()
	properties.
		Set("WorkspaceAlias", f.workspaceAlias).
		Set("WorkspaceARN", f.workspaceARN).
		Set("WorkspaceId", f.workspaceId)

	return properties
}
