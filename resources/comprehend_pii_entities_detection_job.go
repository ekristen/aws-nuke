package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/comprehend"

	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const ComprehendPiiEntititesDetectionJobResource = "ComprehendPiiEntititesDetectionJob"

func init() {
	resource.Register(&resource.Registration{
		Name:   ComprehendPiiEntititesDetectionJobResource,
		Scope:  nuke.Account,
		Lister: &ComprehendPiiEntititesDetectionJobLister{},
	})
}

type ComprehendPiiEntititesDetectionJobLister struct{}

func (l *ComprehendPiiEntititesDetectionJobLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := comprehend.New(opts.Session)

	params := &comprehend.ListPiiEntitiesDetectionJobsInput{}
	resources := make([]resource.Resource, 0)

	for {
		resp, err := svc.ListPiiEntitiesDetectionJobs(params)
		if err != nil {
			return nil, err
		}
		for _, piiEntititesDetectionJob := range resp.PiiEntitiesDetectionJobPropertiesList {
			switch *piiEntititesDetectionJob.JobStatus {
			case "STOPPED", "FAILED", "COMPLETED":
				// if the job has already been stopped, failed, or completed; do not try to stop it again
				continue
			}
			resources = append(resources, &ComprehendPiiEntitiesDetectionJob{
				svc:                      svc,
				piiEntititesDetectionJob: piiEntititesDetectionJob,
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
	svc                      *comprehend.Comprehend
	piiEntititesDetectionJob *comprehend.PiiEntitiesDetectionJobProperties
}

func (ce *ComprehendPiiEntitiesDetectionJob) Remove(_ context.Context) error {
	_, err := ce.svc.StopPiiEntitiesDetectionJob(&comprehend.StopPiiEntitiesDetectionJobInput{
		JobId: ce.piiEntititesDetectionJob.JobId,
	})
	return err
}

func (ce *ComprehendPiiEntitiesDetectionJob) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("JobName", ce.piiEntititesDetectionJob.JobName)
	properties.Set("JobId", ce.piiEntititesDetectionJob.JobId)

	return properties
}

func (ce *ComprehendPiiEntitiesDetectionJob) String() string {
	if ce.piiEntititesDetectionJob.JobName == nil {
		return "Unnamed job"
	} else {
		return *ce.piiEntititesDetectionJob.JobName
	}
}
