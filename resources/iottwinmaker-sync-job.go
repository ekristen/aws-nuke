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

const IoTTwinMakerSyncJobResource = "IoTTwinMakerSyncJob"

func init() {
	registry.Register(&registry.Registration{
		Name:   IoTTwinMakerSyncJobResource,
		Scope:  nuke.Account,
		Lister: &IoTTwinMakerSyncJobLister{},
	})
}

type IoTTwinMakerSyncJobLister struct{}

func (l *IoTTwinMakerSyncJobLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := iottwinmaker.New(opts.Session)
	resources := make([]resource.Resource, 0)

	// Require to have workspaces identifiers to query sync jobs
	workspaceListResponse, err := ListWorkspacesSyncJob(svc)

	if err != nil {
		return nil, err
	}

	for _, workspaceResponse := range workspaceListResponse {
		params := &iottwinmaker.ListSyncJobsInput{
			WorkspaceId: workspaceResponse.WorkspaceId,
			MaxResults:  aws.Int64(25),
		}

		for {
			resp, err := svc.ListSyncJobs(params)
			if err != nil {
				return nil, err
			}

			for _, item := range resp.SyncJobSummaries {
				resources = append(resources, &IoTTwinMakerSyncJob{
					svc:         svc,
					arn:         item.Arn,
					syncSource:  item.SyncSource,
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
func ListWorkspacesSyncJob(svc *iottwinmaker.IoTTwinMaker) ([]*iottwinmaker.WorkspaceSummary, error) {
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

type IoTTwinMakerSyncJob struct {
	svc         *iottwinmaker.IoTTwinMaker
	arn         *string
	syncSource  *string
	WorkspaceID *string
}

func (r *IoTTwinMakerSyncJob) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *IoTTwinMakerSyncJob) Remove(_ context.Context) error {
	_, err := r.svc.DeleteSyncJob(&iottwinmaker.DeleteSyncJobInput{
		SyncSource:  r.syncSource,
		WorkspaceId: r.WorkspaceID,
	})

	return err
}

func (r *IoTTwinMakerSyncJob) String() string {
	return *r.arn
}
