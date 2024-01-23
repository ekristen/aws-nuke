package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lexmodelbuildingservice"

	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const LexIntentResource = "LexIntent"

func init() {
	resource.Register(&resource.Registration{
		Name:   LexIntentResource,
		Scope:  nuke.Account,
		Lister: &LexIntentLister{},
	})
}

type LexIntentLister struct{}

func (l *LexIntentLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := lexmodelbuildingservice.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &lexmodelbuildingservice.GetIntentsInput{
		MaxResults: aws.Int64(20),
	}

	for {
		output, err := svc.GetIntents(params)
		if err != nil {
			return nil, err
		}

		for _, bot := range output.Intents {
			resources = append(resources, &LexIntent{
				svc:  svc,
				name: bot.Name,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type LexIntent struct {
	svc  *lexmodelbuildingservice.LexModelBuildingService
	name *string
}

func (f *LexIntent) Remove(_ context.Context) error {
	_, err := f.svc.DeleteIntent(&lexmodelbuildingservice.DeleteIntentInput{
		Name: f.name,
	})

	return err
}

func (f *LexIntent) String() string {
	return *f.name
}

func (f *LexIntent) Properties() types.Properties {
	properties := types.NewProperties()

	properties.
		Set("Name", f.name)
	return properties
}
