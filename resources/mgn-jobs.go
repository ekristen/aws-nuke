package resources

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/mgn"
	"github.com/aws/aws-sdk-go-v2/service/mgn/types"
	"github.com/aws/smithy-go"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	libtypes "github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const (
	MGNJobResource                      = "MGNJob"
	mgnJobUninitializedAccountException = "UninitializedAccountException"
)

func init() {
	registry.Register(&registry.Registration{
		Name:     MGNJobResource,
		Scope:    nuke.Account,
		Resource: &MGNJob{},
		Lister:   &MGNJobLister{},
	})
}

type MGNJobLister struct{}

func (l *MGNJobLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := mgn.NewFromConfig(*opts.Config)
	resources := make([]resource.Resource, 0)

	params := &mgn.DescribeJobsInput{
		MaxResults: aws.Int32(50),
	}

	for {
		output, err := svc.DescribeJobs(ctx, params)
		if err != nil {
			var apiErr smithy.APIError
			if errors.As(err, &apiErr) && apiErr.ErrorCode() == mgnJobUninitializedAccountException {
				return nil, nil
			}
			return nil, err
		}

		for i := range output.Items {
			job := &output.Items[i]
			mgnJob := &MGNJob{
				svc:         svc,
				job:         job,
				JobID:       job.JobID,
				Arn:         job.Arn,
				Type:        string(job.Type),
				Status:      string(job.Status),
				InitiatedBy: string(job.InitiatedBy),
				Tags:        job.Tags,
			}

			if job.CreationDateTime != nil {
				mgnJob.CreationDateTime = job.CreationDateTime
			}
			if job.EndDateTime != nil {
				mgnJob.EndDateTime = job.EndDateTime
			}

			resources = append(resources, mgnJob)
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type MGNJob struct {
	svc *mgn.Client `description:"-"`
	job *types.Job  `description:"-"`

	// Exposed properties
	JobID            *string           `description:"The unique identifier of the job"`
	Arn              *string           `description:"The ARN of the job"`
	Type             string            `description:"The type of job (LAUNCH, TERMINATE, etc.)"`
	Status           string            `description:"The status of the job"`
	InitiatedBy      string            `description:"Who initiated the job"`
	CreationDateTime *string           `description:"The date and time the job was created"`
	EndDateTime      *string           `description:"The date and time the job ended"`
	Tags             map[string]string `description:"The tags associated with the job"`
}

func (f *MGNJob) Remove(ctx context.Context) error {
	_, err := f.svc.DeleteJob(ctx, &mgn.DeleteJobInput{
		JobID: f.job.JobID,
	})

	return err
}

func (f *MGNJob) Properties() libtypes.Properties {
	return libtypes.NewPropertiesFromStruct(f)
}

func (f *MGNJob) String() string {
	return *f.JobID
}
