package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/gluedatabrew"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const GlueDataBrewRecipeResource = "GlueDataBrewRecipe"

func init() {
	registry.Register(&registry.Registration{
		Name:   GlueDataBrewRecipeResource,
		Scope:  nuke.Account,
		Lister: &GlueDataBrewRecipeLister{},
	})
}

type GlueDataBrewRecipeLister struct{}

func (l *GlueDataBrewRecipeLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := gluedatabrew.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &gluedatabrew.ListRecipesInput{
		MaxResults: aws.Int64(100),
	}

	for {
		output, err := svc.ListRecipes(params)
		if err != nil {
			return nil, err
		}

		for _, recipe := range output.Recipes {
			resources = append(resources, &GlueDataBrewRecipe{
				svc:           svc,
				name:          recipe.Name,
				recipeVersion: recipe.RecipeVersion,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type GlueDataBrewRecipe struct {
	svc           *gluedatabrew.GlueDataBrew
	name          *string
	recipeVersion *string
}

func (f *GlueDataBrewRecipe) Remove(_ context.Context) error {
	_, err := f.svc.DeleteRecipeVersion(&gluedatabrew.DeleteRecipeVersionInput{
		Name:          f.name,
		RecipeVersion: f.recipeVersion,
	})

	return err
}

func (f *GlueDataBrewRecipe) String() string {
	return *f.name
}
