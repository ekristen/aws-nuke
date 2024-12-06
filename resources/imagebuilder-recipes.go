package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/imagebuilder"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const ImageBuilderRecipeResource = "ImageBuilderRecipe"

func init() {
	registry.Register(&registry.Registration{
		Name:     ImageBuilderRecipeResource,
		Scope:    nuke.Account,
		Resource: &ImageBuilderRecipe{},
		Lister:   &ImageBuilderRecipeLister{},
	})
}

type ImageBuilderRecipeLister struct{}

func (l *ImageBuilderRecipeLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := imagebuilder.New(opts.Session)
	params := &imagebuilder.ListImageRecipesInput{}
	resources := make([]resource.Resource, 0)

	for {
		resp, err := svc.ListImageRecipes(params)
		if err != nil {
			return nil, err
		}

		for _, out := range resp.ImageRecipeSummaryList {
			resources = append(resources, &ImageBuilderRecipe{
				svc: svc,
				arn: *out.Arn,
			})
		}

		if resp.NextToken == nil {
			break
		}

		params = &imagebuilder.ListImageRecipesInput{
			NextToken: resp.NextToken,
		}
	}

	return resources, nil
}

type ImageBuilderRecipe struct {
	svc *imagebuilder.Imagebuilder
	arn string
}

func (e *ImageBuilderRecipe) Remove(_ context.Context) error {
	_, err := e.svc.DeleteImageRecipe(&imagebuilder.DeleteImageRecipeInput{
		ImageRecipeArn: &e.arn,
	})
	return err
}

func (e *ImageBuilderRecipe) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("arn", e.arn)
	return properties
}

func (e *ImageBuilderRecipe) String() string {
	return e.arn
}
