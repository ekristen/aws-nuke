package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/imagebuilder"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const ImageBuilderPipelineResource = "ImageBuilderPipeline"

func init() {
	registry.Register(&registry.Registration{
		Name:   ImageBuilderPipelineResource,
		Scope:  nuke.Account,
		Lister: &ImageBuilderPipelineLister{},
	})
}

type ImageBuilderPipelineLister struct{}

func (l *ImageBuilderPipelineLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := imagebuilder.New(opts.Session)
	params := &imagebuilder.ListImagePipelinesInput{}
	resources := make([]resource.Resource, 0)

	for {
		resp, err := svc.ListImagePipelines(params)

		if err != nil {
			return nil, err
		}

		for _, out := range resp.ImagePipelineList {
			resources = append(resources, &ImageBuilderPipeline{
				svc: svc,
				arn: *out.Arn,
			})
		}

		if resp.NextToken == nil {
			break
		}

		params = &imagebuilder.ListImagePipelinesInput{
			NextToken: resp.NextToken,
		}
	}
	return resources, nil
}

type ImageBuilderPipeline struct {
	svc *imagebuilder.Imagebuilder
	arn string
}

func (e *ImageBuilderPipeline) Remove(_ context.Context) error {
	_, err := e.svc.DeleteImagePipeline(&imagebuilder.DeleteImagePipelineInput{
		ImagePipelineArn: &e.arn,
	})
	return err
}

func (e *ImageBuilderPipeline) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("arn", e.arn)
	return properties
}

func (e *ImageBuilderPipeline) String() string {
	return e.arn
}
