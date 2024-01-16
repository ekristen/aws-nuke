package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/comprehend"

	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const ComprehendDominantLanguageDetectionJobResource = "ComprehendDominantLanguageDetectionJob"

func init() {
	resource.Register(resource.Registration{
		Name:   ComprehendDominantLanguageDetectionJobResource,
		Scope:  nuke.Account,
		Lister: &ComprehendDominantLanguageDetectionJobLister{},
	})
}

type ComprehendDominantLanguageDetectionJobLister struct{}

func (l *ComprehendDominantLanguageDetectionJobLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := comprehend.New(opts.Session)

	params := &comprehend.ListDominantLanguageDetectionJobsInput{}
	resources := make([]resource.Resource, 0)

	for {
		resp, err := svc.ListDominantLanguageDetectionJobs(params)
		if err != nil {
			return nil, err
		}
		for _, dominantLanguageDetectionJob := range resp.DominantLanguageDetectionJobPropertiesList {
			resources = append(resources, &ComprehendDominantLanguageDetectionJob{
				svc:                          svc,
				dominantLanguageDetectionJob: dominantLanguageDetectionJob,
			})
		}

		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return resources, nil
}

type ComprehendDominantLanguageDetectionJob struct {
	svc                          *comprehend.Comprehend
	dominantLanguageDetectionJob *comprehend.DominantLanguageDetectionJobProperties
}

func (ce *ComprehendDominantLanguageDetectionJob) Remove(_ context.Context) error {
	_, err := ce.svc.StopDominantLanguageDetectionJob(&comprehend.StopDominantLanguageDetectionJobInput{
		JobId: ce.dominantLanguageDetectionJob.JobId,
	})
	return err
}

func (ce *ComprehendDominantLanguageDetectionJob) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("JobName", ce.dominantLanguageDetectionJob.JobName)
	properties.Set("JobId", ce.dominantLanguageDetectionJob.JobId)

	return properties
}

func (ce *ComprehendDominantLanguageDetectionJob) String() string {
	if ce.dominantLanguageDetectionJob.JobName == nil {
		return "Unnamed job"
	} else {
		return *ce.dominantLanguageDetectionJob.JobName
	}
}
