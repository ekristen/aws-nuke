package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/imagebuilder"

	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const ImageBuilderImageResource = "ImageBuilderImage"

func init() {
	resource.Register(resource.Registration{
		Name:   ImageBuilderImageResource,
		Scope:  nuke.Account,
		Lister: &ImageBuilderImageLister{},
	})
}

type ImageBuilderImageLister struct{}

func (l *ImageBuilderImageLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := imagebuilder.New(opts.Session)
	params := &imagebuilder.ListImagesInput{}
	resources := make([]resource.Resource, 0)

	for {
		resp, err := svc.ListImages(params)
		if err != nil {
			return nil, err
		}

		for _, out := range resp.ImageVersionList {
			resources, err = ImageBuildVersions(svc, out.Arn, resources)
			if err != nil {
				return nil, err
			}
		}

		if resp.NextToken == nil {
			break
		}

		params = &imagebuilder.ListImagesInput{
			NextToken: resp.NextToken,
		}
	}

	return resources, nil
}

func ImageBuildVersions(svc *imagebuilder.Imagebuilder, imageVersionArn *string, resources []resource.Resource) ([]resource.Resource, error) {
	params := &imagebuilder.ListImageBuildVersionsInput{
		ImageVersionArn: imageVersionArn,
	}

	for {
		resp, err := svc.ListImageBuildVersions(params)
		if err != nil {
			return nil, err
		}

		for _, out := range resp.ImageSummaryList {
			resources = append(resources, &ImageBuilderImage{
				svc: svc,
				arn: *out.Arn,
			})
		}

		if resp.NextToken == nil {
			break
		}

		params = &imagebuilder.ListImageBuildVersionsInput{
			ImageVersionArn: imageVersionArn,
			NextToken:       resp.NextToken,
		}
	}
	return resources, nil
}

type ImageBuilderImage struct {
	svc *imagebuilder.Imagebuilder
	arn string
}

func (e *ImageBuilderImage) Remove(_ context.Context) error {
	_, err := e.svc.DeleteImage(&imagebuilder.DeleteImageInput{
		ImageBuildVersionArn: &e.arn,
	})
	return err
}

func (e *ImageBuilderImage) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("arn", e.arn)
	return properties
}

func (e *ImageBuilderImage) String() string {
	return e.arn
}
