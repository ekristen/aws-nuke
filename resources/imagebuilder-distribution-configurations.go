package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/imagebuilder"

	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const ImageBuilderDistributionConfigurationResource = "ImageBuilderDistributionConfiguration"

func init() {
	resource.Register(resource.Registration{
		Name:   ImageBuilderDistributionConfigurationResource,
		Scope:  nuke.Account,
		Lister: &ImageBuilderDistributionConfigurationLister{},
	})
}

type ImageBuilderDistributionConfigurationLister struct{}

func (l *ImageBuilderDistributionConfigurationLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := imagebuilder.New(opts.Session)
	params := &imagebuilder.ListDistributionConfigurationsInput{}
	resources := make([]resource.Resource, 0)

	for {
		resp, err := svc.ListDistributionConfigurations(params)
		if err != nil {
			return nil, err
		}

		for _, out := range resp.DistributionConfigurationSummaryList {
			resources = append(resources, &ImageBuilderDistributionConfiguration{
				svc: svc,
				arn: *out.Arn,
			})
		}

		if resp.NextToken == nil {
			break
		}

		params = &imagebuilder.ListDistributionConfigurationsInput{
			NextToken: resp.NextToken,
		}
	}

	return resources, nil
}

type ImageBuilderDistributionConfiguration struct {
	svc *imagebuilder.Imagebuilder
	arn string
}

func (e *ImageBuilderDistributionConfiguration) Remove(_ context.Context) error {
	_, err := e.svc.DeleteDistributionConfiguration(&imagebuilder.DeleteDistributionConfigurationInput{
		DistributionConfigurationArn: &e.arn,
	})
	return err
}

func (e *ImageBuilderDistributionConfiguration) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("arn", e.arn)
	return properties
}

func (e *ImageBuilderDistributionConfiguration) String() string {
	return e.arn
}
