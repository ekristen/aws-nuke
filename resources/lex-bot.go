package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lexmodelbuildingservice"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const LexBotResource = "LexBot"

func init() {
	registry.Register(&registry.Registration{
		Name:   LexBotResource,
		Scope:  nuke.Account,
		Lister: &LexBotLister{},
	})
}

type LexBotLister struct{}

func (l *LexBotLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := lexmodelbuildingservice.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &lexmodelbuildingservice.GetBotsInput{
		MaxResults: aws.Int64(10),
	}

	for {
		output, err := svc.GetBots(params)
		if err != nil {
			return nil, err
		}

		for _, bot := range output.Bots {
			resources = append(resources, &LexBot{
				svc:    svc,
				name:   bot.Name,
				status: bot.Status,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type LexBot struct {
	svc    *lexmodelbuildingservice.LexModelBuildingService
	name   *string
	status *string
}

func (f *LexBot) Remove(_ context.Context) error {
	_, err := f.svc.DeleteBot(&lexmodelbuildingservice.DeleteBotInput{
		Name: f.name,
	})

	return err
}

func (f *LexBot) String() string {
	return *f.name
}

func (f *LexBot) Properties() types.Properties {
	properties := types.NewProperties()

	properties.
		Set("Name", f.name).
		Set("Status", f.status)
	return properties
}
