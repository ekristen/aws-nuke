package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/imagebuilder"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const ImageBuilderComponentResource = "ImageBuilderComponent"

func init() {
	registry.Register(&registry.Registration{
		Name:   ImageBuilderComponentResource,
		Scope:  nuke.Account,
		Lister: &ImageBuilderComponentLister{},
	})
}

type ImageBuilderComponentLister struct{}

func (l *ImageBuilderComponentLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := imagebuilder.New(opts.Session)
	params := &imagebuilder.ListComponentsInput{}
	resources := make([]resource.Resource, 0)

	for {
		resp, err := svc.ListComponents(params)
		if err != nil {
			return nil, err
		}

		for _, out := range resp.ComponentVersionList {
			resources, err = ListImageBuilderComponentVersions(svc, out.Arn, resources)
			if err != nil {
				return nil, err
			}
		}

		if resp.NextToken == nil {
			break
		}

		params = &imagebuilder.ListComponentsInput{
			NextToken: resp.NextToken,
		}
	}

	return resources, nil
}

func ListImageBuilderComponentVersions(
	svc *imagebuilder.Imagebuilder,
	componentVersionArn *string,
	resources []resource.Resource) ([]resource.Resource, error) {
	params := &imagebuilder.ListComponentBuildVersionsInput{
		ComponentVersionArn: componentVersionArn,
	}

	for {
		resp, err := svc.ListComponentBuildVersions(params)
		if err != nil {
			return nil, err
		}

		for _, out := range resp.ComponentSummaryList {
			resources = append(resources, &ImageBuilderComponent{
				svc: svc,
				arn: *out.Arn,
			})
		}

		if resp.NextToken == nil {
			break
		}

		params = &imagebuilder.ListComponentBuildVersionsInput{
			ComponentVersionArn: componentVersionArn,
			NextToken:           resp.NextToken,
		}
	}
	return resources, nil
}

type ImageBuilderComponent struct {
	svc *imagebuilder.Imagebuilder
	arn string
}

func (e *ImageBuilderComponent) Remove(_ context.Context) error {
	_, err := e.svc.DeleteComponent(&imagebuilder.DeleteComponentInput{
		ComponentBuildVersionArn: &e.arn,
	})
	return err
}

func (e *ImageBuilderComponent) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("arn", e.arn)
	return properties
}

func (e *ImageBuilderComponent) String() string {
	return e.arn
}
