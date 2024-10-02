package resources

import (
	"context"

	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/bedrock"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const BedrockEvaluationJobResource = "BedrockEvaluationJob"

func init() {
	registry.Register(&registry.Registration{
		Name:   BedrockEvaluationJobResource,
		Scope:  nuke.Account,
		Lister: &BedrockEvaluationJobLister{},
	})
}

type BedrockEvaluationJobLister struct{}

func (l *BedrockEvaluationJobLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := bedrock.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &bedrock.ListEvaluationJobsInput{
		MaxResults: aws.Int64(30),
	}

	for {
		resp, err := svc.ListEvaluationJobs(params)
		if err != nil {
			return nil, err
		}

		for _, jobSummary := range resp.JobSummaries {
			resources = append(resources, &BedrockEvaluationJob{
				svc:     svc,
				jobName: jobSummary.JobName,
				jobArn:  jobSummary.JobArn,
				status:  jobSummary.status,
			})
		}

		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return resources, nil
}

type BedrockEvaluationJob struct {
	svc     *bedrock.Bedrock
	jobName *string
	jobArn  *string
	status  *string
}

func (f *BedrockEvaluationJob) Remove(_ context.Context) error {
	_, err := f.svc.StopEvaluationJob(&bedrock.StopEvaluationJobInput{
		JobIdentifier: f.jobArn,
	})

	return err
}

func (f *BedrockEvaluationJob) String() string {
	return *f.jobName
}

func (f *BedrockEvaluationJob) Properties() types.Properties {
	properties := types.NewProperties()
	properties.
		Set("JobName", f.jobName).
		Set("Status", f.status)
	return properties
}

func (f *BedrockEvaluationJob) Filter() error {
	if *f.status != bedrock.EvaluationJobStatusInProgress {
		// May be completed, failed, stopping or stopped
		return fmt.Errorf("already stopped")
	}
	return nil
}
