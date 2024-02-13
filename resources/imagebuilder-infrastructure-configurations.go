package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/imagebuilder"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const ImageBuilderInfrastructureConfigurationResource = "ImageBuilderInfrastructureConfiguration"

func init() {
	registry.Register(&registry.Registration{
		Name:   ImageBuilderInfrastructureConfigurationResource,
		Scope:  nuke.Account,
		Lister: &ImageBuilderInfrastructureConfigurationLister{},
	})
}

type ImageBuilderInfrastructureConfigurationLister struct{}

func (l *ImageBuilderInfrastructureConfigurationLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := imagebuilder.New(opts.Session)
	params := &imagebuilder.ListInfrastructureConfigurationsInput{}
	resources := make([]resource.Resource, 0)

	for {
		resp, err := svc.ListInfrastructureConfigurations(params)
		if err != nil {
			return nil, err
		}

		for _, out := range resp.InfrastructureConfigurationSummaryList {
			resources = append(resources, &ImageBuilderInfrastructureConfiguration{
				svc: svc,
				arn: *out.Arn,
			})
		}

		if resp.NextToken == nil {
			break
		}

		params = &imagebuilder.ListInfrastructureConfigurationsInput{
			NextToken: resp.NextToken,
		}
	}

	return resources, nil
}

type ImageBuilderInfrastructureConfiguration struct {
	svc *imagebuilder.Imagebuilder
	arn string
}

func (e *ImageBuilderInfrastructureConfiguration) Remove(_ context.Context) error {
	_, err := e.svc.DeleteInfrastructureConfiguration(&imagebuilder.DeleteInfrastructureConfigurationInput{
		InfrastructureConfigurationArn: &e.arn,
	})
	return err
}

func (e *ImageBuilderInfrastructureConfiguration) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("arn", e.arn)
	return properties
}

func (e *ImageBuilderInfrastructureConfiguration) String() string {
	return e.arn
}
