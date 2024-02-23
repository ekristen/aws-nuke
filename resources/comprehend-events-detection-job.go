package resources

import (
	"context"

	"github.com/gotidy/ptr"

	"github.com/aws/aws-sdk-go/service/comprehend"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const ComprehendEventsDetectionJobResource = "ComprehendEventsDetectionJob"

func init() {
	registry.Register(&registry.Registration{
		Name:   ComprehendEventsDetectionJobResource,
		Scope:  nuke.Account,
		Lister: &ComprehendEventsDetectionJobLister{},
	})
}

type ComprehendEventsDetectionJobLister struct{}

func (l *ComprehendEventsDetectionJobLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := comprehend.New(opts.Session)

	params := &comprehend.ListEventsDetectionJobsInput{}
	resources := make([]resource.Resource, 0)

	for {
		resp, err := svc.ListEventsDetectionJobs(params)
		if err != nil {
			return nil, err
		}

		for _, eventsDetectionJob := range resp.EventsDetectionJobPropertiesList {
			switch ptr.ToString(eventsDetectionJob.JobStatus) {
			case comprehend.JobStatusStopped, comprehend.JobStatusFailed, comprehend.JobStatusCompleted:
				// if the job has already been stopped, failed, or completed; do not try to stop it again
				continue
			}
			resources = append(resources, &ComprehendEventsDetectionJob{
				svc:                svc,
				eventsDetectionJob: eventsDetectionJob,
			})
		}

		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return resources, nil
}

type ComprehendEventsDetectionJob struct {
	svc                *comprehend.Comprehend
	eventsDetectionJob *comprehend.EventsDetectionJobProperties
}

func (ce *ComprehendEventsDetectionJob) Remove(_ context.Context) error {
	_, err := ce.svc.StopEventsDetectionJob(&comprehend.StopEventsDetectionJobInput{
		JobId: ce.eventsDetectionJob.JobId,
	})
	return err
}

func (ce *ComprehendEventsDetectionJob) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("JobName", ce.eventsDetectionJob.JobName)
	properties.Set("JobId", ce.eventsDetectionJob.JobId)

	return properties
}

func (ce *ComprehendEventsDetectionJob) String() string {
	if ce.eventsDetectionJob.JobName == nil {
		return ComprehendUnnamedJob
	} else {
		return ptr.ToString(ce.eventsDetectionJob.JobName)
	}
}
