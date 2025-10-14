package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"         //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/iot" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const IoTJobResource = "IoTJob"

func init() {
	registry.Register(&registry.Registration{
		Name:     IoTJobResource,
		Scope:    nuke.Account,
		Resource: &IoTJob{},
		Lister:   &IoTJobLister{},
	})
}

type IoTJobLister struct{}

func (l *IoTJobLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := iot.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &iot.ListJobsInput{
		MaxResults: aws.Int64(100),
		Status:     aws.String("IN_PROGRESS"),
	}
	for {
		output, err := svc.ListJobs(params)
		if err != nil {
			return nil, err
		}

		for _, job := range output.Jobs {
			resources = append(resources, &IoTJob{
				svc:    svc,
				ID:     job.JobId,
				status: job.Status,
			})
		}
		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type IoTJob struct {
	svc    *iot.IoT
	ID     *string
	status *string
}

func (f *IoTJob) Remove(_ context.Context) error {
	_, err := f.svc.CancelJob(&iot.CancelJobInput{
		JobId: f.ID,
	})

	return err
}

func (f *IoTJob) String() string {
	return *f.ID
}
