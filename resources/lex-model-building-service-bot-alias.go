package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lexmodelbuildingservice"

	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const LexModelBuildingServiceBotAliasResource = "LexModelBuildingServiceBotAlias"

func init() {
	resource.Register(resource.Registration{
		Name:   LexModelBuildingServiceBotAliasResource,
		Scope:  nuke.Account,
		Lister: &LexModelBuildingServiceBotAliasLister{},
	})
}

type LexModelBuildingServiceBotAliasLister struct{}

func (l *LexModelBuildingServiceBotAliasLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := lexmodelbuildingservice.New(opts.Session)
	resources := make([]resource.Resource, 0)

	botParams := &lexmodelbuildingservice.GetBotsInput{
		MaxResults: aws.Int64(10),
	}

	for {
		botResp, err := svc.GetBots(botParams)
		if err != nil {
			return nil, err
		}

		for _, bot := range botResp.Bots {
			for {
				aliasParams := &lexmodelbuildingservice.GetBotAliasesInput{
					BotName:    bot.Name,
					MaxResults: aws.Int64(10),
				}
				aliasResp, err := svc.GetBotAliases(aliasParams)
				if err != nil {
					continue
				}

				for _, alias := range aliasResp.BotAliases {
					resources = append(resources, &LexModelBuildingServiceBotAlias{
						svc:      svc,
						name:     alias.Name,
						checksum: alias.Checksum,
						botName:  bot.Name,
					})
				}

				if aliasResp.NextToken == nil {
					break
				}
				aliasParams.NextToken = aliasResp.NextToken
			}
		}

		if botResp.NextToken == nil {
			break
		}

		botParams.NextToken = botResp.NextToken
	}

	return resources, nil
}

type LexModelBuildingServiceBotAlias struct {
	svc      *lexmodelbuildingservice.LexModelBuildingService
	name     *string
	checksum *string
	botName  *string
}

func (f *LexModelBuildingServiceBotAlias) Remove(_ context.Context) error {
	params := &lexmodelbuildingservice.DeleteBotAliasInput{
		BotName: f.botName,
		Name:    f.name,
	}
	_, err := f.svc.DeleteBotAlias(params)
	return err
}

func (f *LexModelBuildingServiceBotAlias) String() string {
	return *f.name
}

func (f *LexModelBuildingServiceBotAlias) Properties() types.Properties {
	properties := types.NewProperties()

	properties.Set("Name", f.name)
	properties.Set("BotName", f.botName)
	properties.Set("Checksum", f.checksum)
	return properties
}
