package resources

import (
	"context"

	"github.com/gotidy/ptr"

	"github.com/aws/aws-sdk-go/service/comprehend"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const ComprehendSentimentDetectionJobResource = "ComprehendSentimentDetectionJob"

func init() {
	registry.Register(&registry.Registration{
		Name:     ComprehendSentimentDetectionJobResource,
		Scope:    nuke.Account,
		Resource: &ComprehendSentimentDetectionJob{},
		Lister:   &ComprehendSentimentDetectionJobLister{},
	})
}

type ComprehendSentimentDetectionJobLister struct{}

func (l *ComprehendSentimentDetectionJobLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := comprehend.New(opts.Session)

	params := &comprehend.ListSentimentDetectionJobsInput{}
	resources := make([]resource.Resource, 0)

	for {
		resp, err := svc.ListSentimentDetectionJobs(params)
		if err != nil {
			return nil, err
		}
		for _, sentimentDetectionJob := range resp.SentimentDetectionJobPropertiesList {
			switch *sentimentDetectionJob.JobStatus {
			case "STOPPED", "FAILED", "COMPLETED":
				// if the job has already been stopped, failed, or completed; do not try to stop it again
				continue
			}
			resources = append(resources, &ComprehendSentimentDetectionJob{
				svc:                   svc,
				sentimentDetectionJob: sentimentDetectionJob,
			})
		}

		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return resources, nil
}

type ComprehendSentimentDetectionJob struct {
	svc                   *comprehend.Comprehend
	sentimentDetectionJob *comprehend.SentimentDetectionJobProperties
}

func (ce *ComprehendSentimentDetectionJob) Remove(_ context.Context) error {
	_, err := ce.svc.StopSentimentDetectionJob(&comprehend.StopSentimentDetectionJobInput{
		JobId: ce.sentimentDetectionJob.JobId,
	})
	return err
}

func (ce *ComprehendSentimentDetectionJob) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("JobName", ce.sentimentDetectionJob.JobName)
	properties.Set("JobId", ce.sentimentDetectionJob.JobId)

	return properties
}

func (ce *ComprehendSentimentDetectionJob) String() string {
	if ce.sentimentDetectionJob.JobName == nil {
		return ComprehendUnnamedJob
	} else {
		return ptr.ToString(ce.sentimentDetectionJob.JobName)
	}
}
