package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/managedgrafana"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const AMGWorkspaceResource = "AMGWorkspace"

func init() {
	registry.Register(&registry.Registration{
		Name:     AMGWorkspaceResource,
		Scope:    nuke.Account,
		Resource: &AMGWorkspace{},
		Lister:   &AMGWorkspaceLister{},
	})
}

type AMGWorkspaceLister struct{}

func (l *AMGWorkspaceLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := managedgrafana.New(opts.Session)
	resources := make([]resource.Resource, 0)

	var amgWorkspaces []*managedgrafana.WorkspaceSummary
	err := svc.ListWorkspacesPages(
		&managedgrafana.ListWorkspacesInput{},
		func(page *managedgrafana.ListWorkspacesOutput, lastPage bool) bool {
			amgWorkspaces = append(amgWorkspaces, page.Workspaces...)
			return true
		},
	)
	if err != nil {
		return nil, err
	}

	for _, ws := range amgWorkspaces {
		resources = append(resources, &AMGWorkspace{
			svc:  svc,
			id:   ws.Id,
			name: ws.Name,
		})
	}

	return resources, nil
}

type AMGWorkspace struct {
	svc  *managedgrafana.ManagedGrafana
	id   *string
	name *string
}

func (f *AMGWorkspace) Remove(_ context.Context) error {
	_, err := f.svc.DeleteWorkspace(&managedgrafana.DeleteWorkspaceInput{
		WorkspaceId: f.id,
	})

	return err
}

func (f *AMGWorkspace) Properties() types.Properties {
	properties := types.NewProperties()
	properties.
		Set("WorkspaceId", f.id).
		Set("WorkspaceName", f.name)

	return properties
}
