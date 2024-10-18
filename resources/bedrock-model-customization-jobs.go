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
			resources = append(resources, &BedrockModelCustomizationJob{
				svc:             svc,
				customModelName: modelCustomizationJobSummary.CustomModelName,
				customModelArn:  modelCustomizationJobSummary.CustomModelArn,
				status:          modelCustomizationJobSummary.Status,
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
	svc             *bedrock.Bedrock
	customModelName *string
	customModelArn  *string
	status          *string
}

func (f *BedrockModelCustomizationJob) Remove(_ context.Context) error {
	_, err := f.svc.StopModelCustomizationJob(&bedrock.StopModelCustomizationJobInput{
		JobIdentifier: f.customModelName,
	})

	return err
}

func (f *BedrockModelCustomizationJob) String() string {
	return *f.customModelName
}

func (f *BedrockModelCustomizationJob) Properties() types.Properties {
	properties := types.NewProperties()
	properties.
		Set("CustomModelName", f.customModelName).
		Set("Status", f.status)
	return properties
}

func (f *BedrockModelCustomizationJob) Filter() error {
	if *f.status != bedrock.ModelCustomizationJobStatusInProgress {
		// May be completed, failed, stopping or stopped
		return fmt.Errorf("already stopped")
	}
	return nil
}
