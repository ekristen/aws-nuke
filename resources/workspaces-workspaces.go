package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"                //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/workspaces" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const WorkSpacesWorkspaceResource = "WorkSpacesWorkspace"

func init() {
	registry.Register(&registry.Registration{
		Name:     WorkSpacesWorkspaceResource,
		Scope:    nuke.Account,
		Resource: &WorkSpacesWorkspace{},
		Lister:   &WorkSpacesWorkspaceLister{},
	})
}

type WorkSpacesWorkspaceLister struct{}

func (l *WorkSpacesWorkspaceLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := workspaces.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &workspaces.DescribeWorkspacesInput{
		Limit: aws.Int64(25),
	}

	for {
		output, err := svc.DescribeWorkspaces(params)
		if err != nil {
			return nil, err
		}

		for _, workspace := range output.Workspaces {
			resources = append(resources, &WorkSpacesWorkspace{
				svc:         svc,
				workspaceID: workspace.WorkspaceId,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type WorkSpacesWorkspace struct {
	svc         *workspaces.WorkSpaces
	workspaceID *string
}

func (f *WorkSpacesWorkspace) Remove(_ context.Context) error {
	stopRequest := &workspaces.StopRequest{
		WorkspaceId: f.workspaceID,
	}
	terminateRequest := &workspaces.TerminateRequest{
		WorkspaceId: f.workspaceID,
	}

	_, err := f.svc.StopWorkspaces(&workspaces.StopWorkspacesInput{
		StopWorkspaceRequests: []*workspaces.StopRequest{stopRequest},
	})
	if err != nil {
		return err
	}

	_, err = f.svc.TerminateWorkspaces(&workspaces.TerminateWorkspacesInput{
		TerminateWorkspaceRequests: []*workspaces.TerminateRequest{terminateRequest},
	})

	return err
}

func (f *WorkSpacesWorkspace) String() string {
	return *f.workspaceID
}
