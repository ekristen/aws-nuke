package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/comprehend"

	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const ComprehendKeyPhrasesDetectionJobResource = "ComprehendKeyPhrasesDetectionJob"

func init() {
	resource.Register(&resource.Registration{
		Name:   ComprehendKeyPhrasesDetectionJobResource,
		Scope:  nuke.Account,
		Lister: &ComprehendKeyPhrasesDetectionJobLister{},
	})
}

type ComprehendKeyPhrasesDetectionJobLister struct{}

func (l *ComprehendKeyPhrasesDetectionJobLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := comprehend.New(opts.Session)

	params := &comprehend.ListKeyPhrasesDetectionJobsInput{}
	resources := make([]resource.Resource, 0)

	for {
		resp, err := svc.ListKeyPhrasesDetectionJobs(params)
		if err != nil {
			return nil, err
		}
		for _, keyPhrasesDetectionJob := range resp.KeyPhrasesDetectionJobPropertiesList {
			switch *keyPhrasesDetectionJob.JobStatus {
			case "STOPPED", "FAILED", "COMPLETED":
				// if the job has already been stopped, failed, or completed; do not try to stop it again
				continue
			}
			resources = append(resources, &ComprehendKeyPhrasesDetectionJob{
				svc:                    svc,
				keyPhrasesDetectionJob: keyPhrasesDetectionJob,
			})
		}

		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return resources, nil
}

type ComprehendKeyPhrasesDetectionJob struct {
	svc                    *comprehend.Comprehend
	keyPhrasesDetectionJob *comprehend.KeyPhrasesDetectionJobProperties
}

func (ce *ComprehendKeyPhrasesDetectionJob) Remove(_ context.Context) error {
	_, err := ce.svc.StopKeyPhrasesDetectionJob(&comprehend.StopKeyPhrasesDetectionJobInput{
		JobId: ce.keyPhrasesDetectionJob.JobId,
	})
	return err
}

func (ce *ComprehendKeyPhrasesDetectionJob) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("JobName", ce.keyPhrasesDetectionJob.JobName)
	properties.Set("JobId", ce.keyPhrasesDetectionJob.JobId)

	return properties
}

func (ce *ComprehendKeyPhrasesDetectionJob) String() string {
	if ce.keyPhrasesDetectionJob.JobName == nil {
		return "Unnamed job"
	} else {
		return *ce.keyPhrasesDetectionJob.JobName
	}
}
