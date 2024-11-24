package resources

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/bedrock"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
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
			tagResp, err := svc.ListTagsForResource(
				&bedrock.ListTagsForResourceInput{
					ResourceARN: jobSummary.JobArn,
				})
			if err != nil {
				return nil, err
			}

			resources = append(resources, &BedrockEvaluationJob{
				svc:    svc,
				Arn:    jobSummary.JobArn,
				Name:   jobSummary.JobName,
				Status: jobSummary.Status,
				Tags:   tagResp.Tags,
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
	svc    *bedrock.Bedrock
	Arn    *string
	Name   *string
	Status *string
	Tags   []*bedrock.Tag
}

func (r *BedrockEvaluationJob) Remove(_ context.Context) error {
	// We cannot delete an evaluation job from API, only stop it
	// Deletion seems to be possible in console only for now
	_, err := r.svc.StopEvaluationJob(&bedrock.StopEvaluationJobInput{
		JobIdentifier: r.Arn,
	})

	return err
}

func (r *BedrockEvaluationJob) String() string {
	return *r.Name
}

func (r *BedrockEvaluationJob) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *BedrockEvaluationJob) Filter() error {
	if *r.Status != bedrock.EvaluationJobStatusInProgress {
		// May be completed, failed, stopping or stopped
		return fmt.Errorf("already stopped")
	}
	return nil
}
