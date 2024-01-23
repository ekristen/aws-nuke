package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/glue"

	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const GlueJobResource = "GlueJob"

func init() {
	resource.Register(&resource.Registration{
		Name:   GlueJobResource,
		Scope:  nuke.Account,
		Lister: &GlueJobLister{},
	})
}

type GlueJobLister struct{}

func (l *GlueJobLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := glue.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &glue.GetJobsInput{
		MaxResults: aws.Int64(100),
	}

	for {
		output, err := svc.GetJobs(params)
		if err != nil {
			return nil, err
		}

		for _, job := range output.Jobs {
			resources = append(resources, &GlueJob{
				svc:     svc,
				jobName: job.Name,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type GlueJob struct {
	svc     *glue.Glue
	jobName *string
}

func (f *GlueJob) Remove(_ context.Context) error {
	_, err := f.svc.DeleteJob(&glue.DeleteJobInput{
		JobName: f.jobName,
	})

	return err
}

func (f *GlueJob) String() string {
	return *f.jobName
}
