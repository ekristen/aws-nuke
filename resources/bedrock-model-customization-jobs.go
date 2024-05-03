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

const BedrockModelCustomizationJobResource = "BedrockModelCustomizationJob"

func init() {
	registry.Register(&registry.Registration{
		Name:   BedrockModelCustomizationJobResource,
		Scope:  nuke.Account,
		Lister: &BedrockModelCustomizationJobLister{},
	})
}

type BedrockModelCustomizationJobLister struct{}

func (l *BedrockModelCustomizationJobLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := bedrock.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &bedrock.ListModelCustomizationJobsInput{
		MaxResults: aws.Int64(30),
	}

	for {
		resp, err := svc.ListModelCustomizationJobs(params)
		if err != nil {
			return nil, err
		}

		for _, modelCustomizationJobSummary := range resp.ModelCustomizationJobSummaries {
			tagResp, err := svc.ListTagsForResource(
				&bedrock.ListTagsForResourceInput{
					ResourceARN: modelCustomizationJobSummary.JobArn,
				})
			if err != nil {
				return nil, err
			}
			resources = append(resources, &BedrockModelCustomizationJob{
				svc:       svc,
				Arn:       modelCustomizationJobSummary.JobArn,
				JobName:   modelCustomizationJobSummary.JobName,
				ModelName: modelCustomizationJobSummary.CustomModelName,
				Status:    modelCustomizationJobSummary.Status,
				Tags:      tagResp.Tags,
			})
		}

		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return resources, nil
}

type BedrockModelCustomizationJob struct {
	svc       *bedrock.Bedrock
	Arn       *string
	ModelName *string
	JobName   *string
	Status    *string
	Tags      []*bedrock.Tag
}

func (r *BedrockModelCustomizationJob) Remove(_ context.Context) error {
	_, err := r.svc.StopModelCustomizationJob(&bedrock.StopModelCustomizationJobInput{
		JobIdentifier: r.Arn,
	})

	return err
}

func (r *BedrockModelCustomizationJob) String() string {
	return *r.JobName
}

func (r *BedrockModelCustomizationJob) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *BedrockModelCustomizationJob) Filter() error {
	if *r.Status != bedrock.ModelCustomizationJobStatusInProgress {
		// May be completed, failed, stopping or stopped
		return fmt.Errorf("already stopped")
	}
	return nil
}
