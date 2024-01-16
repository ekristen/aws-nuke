package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/gluedatabrew"

	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const GlueDataBrewRulesetsResource = "GlueDataBrewRulesets"

func init() {
	resource.Register(resource.Registration{
		Name:   GlueDataBrewRulesetsResource,
		Scope:  nuke.Account,
		Lister: &GlueDataBrewRulesetsLister{},
	})
}

type GlueDataBrewRulesetsLister struct{}

func (l *GlueDataBrewRulesetsLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := gluedatabrew.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &gluedatabrew.ListRulesetsInput{
		MaxResults: aws.Int64(100),
	}

	for {
		output, err := svc.ListRulesets(params)
		if err != nil {
			return nil, err
		}

		for _, ruleset := range output.Rulesets {
			resources = append(resources, &GlueDataBrewRulesets{
				svc:  svc,
				name: ruleset.Name,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type GlueDataBrewRulesets struct {
	svc  *gluedatabrew.GlueDataBrew
	name *string
}

func (f *GlueDataBrewRulesets) Remove(_ context.Context) error {
	_, err := f.svc.DeleteRuleset(&gluedatabrew.DeleteRulesetInput{
		Name: f.name,
	})

	return err
}

func (f *GlueDataBrewRulesets) String() string {
	return *f.name
}
