package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/bedrock"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const ProvisionedModelThroughputResource = "ProvisionedModelThroughput"

func init() {
	registry.Register(&registry.Registration{
		Name:   ProvisionedModelThroughputResource,
		Scope:  nuke.Account,
		Lister: &ProvisionedModelThroughputLister{},
	})
}

type ProvisionedModelThroughputLister struct{}

func (l *ProvisionedModelThroughputLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := bedrock.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &bedrock.ListProvisionedModelThroughputsInput{
		MaxResults: aws.Int64(30),
	}

	for {
		resp, err := svc.ListProvisionedModelThroughputs(params)
		if err != nil {
			return nil, err
		}

		for _, provisionedModelSummary := range resp.ProvisionedModelSummaries {
			resources = append(resources, &ProvisionedModelThroughput{
				svc:                  svc,
				provisionedModelArn:  provisionedModelSummary.ProvisionedModelArn,
				provisionedModelName: provisionedModelSummary.ProvisionedModelName,
				status:               provisionedModelSummary.Status,
			})
		}

		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return resources, nil
}

type ProvisionedModelThroughput struct {
	svc                  *bedrock.Bedrock
	provisionedModelArn  *string
	provisionedModelName *string
	status               *string
}

func (f *ProvisionedModelThroughput) Remove(_ context.Context) error {
	_, err := f.svc.DeleteProvisionedModelThroughput(&bedrock.DeleteProvisionedModelThroughputInput{
		ProvisionedModelId: f.provisionedModelArn,
	})

	return err
}

func (f *ProvisionedModelThroughput) String() string {
	return *f.provisionedModelName
}

func (f *ProvisionedModelThroughput) Properties() types.Properties {
	properties := types.NewProperties()
	properties.
		Set("ProvisionedModelArn", f.provisionedModelArn).
		Set("ProvisionedModelName", f.provisionedModelName).
		Set("Status", f.status)
	return properties
}
