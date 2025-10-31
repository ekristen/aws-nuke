package resources

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/emrserverless"
	"github.com/aws/aws-sdk-go-v2/service/emrserverless/types"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	libtypes "github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const EMRServerlessJobRunResource = "EMRServerlessJobRun"

func init() {
	registry.Register(&registry.Registration{
		Name:     EMRServerlessJobRunResource,
		Scope:    nuke.Account,
		Resource: &EMRServerlessJobRun{},
		Lister:   &EMRServerlessJobRunLister{},
	})
}

type EMRServerlessJobRunLister struct{}

func (l *EMRServerlessJobRunLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := emrserverless.NewFromConfig(*opts.Config)

	var resources []resource.Resource

	appsParams := &emrserverless.ListApplicationsInput{
		MaxResults: aws.Int32(50),
	}

	appsPaginator := emrserverless.NewListApplicationsPaginator(svc, appsParams)

	for appsPaginator.HasMorePages() {
		appsPage, err := appsPaginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, app := range appsPage.Applications {
			jobParams := &emrserverless.ListJobRunsInput{
				ApplicationId: app.Id,
				MaxResults:    aws.Int32(50),
			}

			jobPaginator := emrserverless.NewListJobRunsPaginator(svc, jobParams)

			for jobPaginator.HasMorePages() {
				jobPage, err := jobPaginator.NextPage(ctx)
				if err != nil {
					return nil, err
				}

				for _, jobRun := range jobPage.JobRuns {
					if isJobRunCancellable(jobRun.State) {
						detailsOutput, err := svc.GetJobRun(ctx, &emrserverless.GetJobRunInput{
							ApplicationId: app.Id,
							JobRunId:      jobRun.Id,
						})
						if err != nil {
							return nil, err
						}

						resources = append(resources, &EMRServerlessJobRun{
							svc:             svc,
							ApplicationID:   app.Id,
							ApplicationName: app.Name,
							JobRunID:        jobRun.Id,
							Name:            jobRun.Name,
							ARN:             jobRun.Arn,
							State:           jobRun.State,
							CreatedAt:       jobRun.CreatedAt,
							UpdatedAt:       jobRun.UpdatedAt,
							Tags:            detailsOutput.JobRun.Tags,
						})
					}
				}
			}
		}
	}

	return resources, nil
}

// isJobRunCancellable returns true if the job run is in a state that can be cancelled
func isJobRunCancellable(state types.JobRunState) bool {
	switch state {
	case types.JobRunStateSubmitted, types.JobRunStatePending,
		types.JobRunStateScheduled, types.JobRunStateRunning:
		return true
	default:
		return false
	}
}

type EMRServerlessJobRun struct {
	svc             *emrserverless.Client
	ApplicationID   *string
	ApplicationName *string
	JobRunID        *string
	Name            *string
	ARN             *string
	State           types.JobRunState
	CreatedAt       *time.Time
	UpdatedAt       *time.Time
	Tags            map[string]string
}

func (r *EMRServerlessJobRun) Remove(ctx context.Context) error {
	_, err := r.svc.CancelJobRun(ctx, &emrserverless.CancelJobRunInput{
		ApplicationId: r.ApplicationID,
		JobRunId:      r.JobRunID,
	})

	return err
}

func (r *EMRServerlessJobRun) Filter() error {
	if !isJobRunCancellable(r.State) {
		return fmt.Errorf("job run is not in a cancellable state: %s", r.State)
	}
	return nil
}

func (r *EMRServerlessJobRun) Properties() libtypes.Properties {
	properties := libtypes.NewProperties()
	properties.
		Set("ApplicationID", r.ApplicationID).
		Set("ApplicationName", r.ApplicationName).
		Set("JobRunID", r.JobRunID).
		Set("Name", r.Name).
		Set("ARN", r.ARN).
		Set("State", string(r.State))

	if r.CreatedAt != nil {
		properties.Set("CreatedAt", r.CreatedAt.Format(time.RFC3339))
	}
	if r.UpdatedAt != nil {
		properties.Set("UpdatedAt", r.UpdatedAt.Format(time.RFC3339))
	}

	for key, val := range r.Tags {
		properties.SetTag(&key, &val)
	}

	return properties
}

func (r *EMRServerlessJobRun) String() string {
	return *r.JobRunID
}
