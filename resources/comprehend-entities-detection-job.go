package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/comprehend"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const ComprehendEntitiesDetectionJobResource = "ComprehendEntitiesDetectionJob"

func init() {
	registry.Register(&registry.Registration{
		Name:   ComprehendEntitiesDetectionJobResource,
		Scope:  nuke.Account,
		Lister: &ComprehendEntitiesDetectionJobLister{},
	})
}

type ComprehendEntitiesDetectionJobLister struct{}

func (l *ComprehendEntitiesDetectionJobLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := comprehend.New(opts.Session)

	params := &comprehend.ListEntitiesDetectionJobsInput{}
	resources := make([]resource.Resource, 0)

	for {
		resp, err := svc.ListEntitiesDetectionJobs(params)
		if err != nil {
			return nil, err
		}
		for _, entitiesDetectionJob := range resp.EntitiesDetectionJobPropertiesList {
			switch *entitiesDetectionJob.JobStatus {
			case "STOPPED", "FAILED", "COMPLETED":
				// if the job has already been stopped, failed, or completed; do not try to stop it again
				continue
			}
			resources = append(resources, &ComprehendEntitiesDetectionJob{
				svc:                  svc,
				entitiesDetectionJob: entitiesDetectionJob,
			})
		}

		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return resources, nil
}

type ComprehendEntitiesDetectionJob struct {
	svc                  *comprehend.Comprehend
	entitiesDetectionJob *comprehend.EntitiesDetectionJobProperties
}

func (ce *ComprehendEntitiesDetectionJob) Remove(_ context.Context) error {
	_, err := ce.svc.StopEntitiesDetectionJob(&comprehend.StopEntitiesDetectionJobInput{
		JobId: ce.entitiesDetectionJob.JobId,
	})
	return err
}

func (ce *ComprehendEntitiesDetectionJob) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("JobName", ce.entitiesDetectionJob.JobName)
	properties.Set("JobId", ce.entitiesDetectionJob.JobId)

	return properties
}

func (ce *ComprehendEntitiesDetectionJob) String() string {
	if ce.entitiesDetectionJob.JobName == nil {
		return "Unnamed job"
	} else {
		return *ce.entitiesDetectionJob.JobName
	}
}
