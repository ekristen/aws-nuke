package resources

import (
	"context"

	"github.com/gotidy/ptr"

	"github.com/aws/aws-sdk-go/service/comprehend" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const ComprehendPiiEntitiesDetectionJobResource = "ComprehendPiiEntitiesDetectionJob"

func init() {
	registry.Register(&registry.Registration{
		Name:     ComprehendPiiEntitiesDetectionJobResource,
		Scope:    nuke.Account,
		Resource: &ComprehendPiiEntitiesDetectionJob{},
		Lister:   &ComprehendPiiEntitiesDetectionJobLister{},
		DeprecatedAliases: []string{
			"ComprehendPiiEntititesDetectionJob",
		},
	})
}

type ComprehendPiiEntitiesDetectionJobLister struct{}

func (l *ComprehendPiiEntitiesDetectionJobLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := comprehend.New(opts.Session)

	params := &comprehend.ListPiiEntitiesDetectionJobsInput{}
	resources := make([]resource.Resource, 0)

	for {
		resp, err := svc.ListPiiEntitiesDetectionJobs(params)
		if err != nil {
			return nil, err
		}
		for _, piiEntitiesDetectionJob := range resp.PiiEntitiesDetectionJobPropertiesList {
			switch *piiEntitiesDetectionJob.JobStatus {
			case comprehend.JobStatusStopped, comprehend.JobStatusFailed, comprehend.JobStatusCompleted:
				// if the job has already been stopped, failed, or completed; do not try to stop it again
				continue
			}
			resources = append(resources, &ComprehendPiiEntitiesDetectionJob{
				svc:                     svc,
				piiEntitiesDetectionJob: piiEntitiesDetectionJob,
			})
		}

		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return resources, nil
}

type ComprehendPiiEntitiesDetectionJob struct {
	svc                     *comprehend.Comprehend
	piiEntitiesDetectionJob *comprehend.PiiEntitiesDetectionJobProperties
}

func (ce *ComprehendPiiEntitiesDetectionJob) Remove(_ context.Context) error {
	_, err := ce.svc.StopPiiEntitiesDetectionJob(&comprehend.StopPiiEntitiesDetectionJobInput{
		JobId: ce.piiEntitiesDetectionJob.JobId,
	})
	return err
}

func (ce *ComprehendPiiEntitiesDetectionJob) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("JobName", ce.piiEntitiesDetectionJob.JobName)
	properties.Set("JobId", ce.piiEntitiesDetectionJob.JobId)

	return properties
}

func (ce *ComprehendPiiEntitiesDetectionJob) String() string {
	if ce.piiEntitiesDetectionJob.JobName == nil {
		return ComprehendUnnamedJob
	} else {
		return ptr.ToString(ce.piiEntitiesDetectionJob.JobName)
	}
}
