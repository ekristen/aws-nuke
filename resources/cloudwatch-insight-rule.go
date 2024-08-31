package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/aws/aws-sdk-go/service/cloudwatch"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const CloudWatchInsightRuleResource = "CloudWatchInsightRule"

func init() {
	registry.Register(&registry.Registration{
		Name:   CloudWatchInsightRuleResource,
		Scope:  nuke.Account,
		Lister: &CloudWatchInsightRuleLister{},
	})
}

type CloudWatchInsightRuleLister struct{}

func (l *CloudWatchInsightRuleLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := cloudwatch.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &cloudwatch.DescribeInsightRulesInput{
		MaxResults: aws.Int64(25),
	}

	for {
		output, err := svc.DescribeInsightRules(params)
		if err != nil {
			return nil, err
		}

		for _, rules := range output.InsightRules {
			resources = append(resources, &CloudWatchInsightRule{
				svc:  svc,
				name: rules.Name,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type CloudWatchInsightRule struct {
	svc  *cloudwatch.CloudWatch
	name *string
}

func (f *CloudWatchInsightRule) Remove(_ context.Context) error {
	_, err := f.svc.DeleteInsightRules(&cloudwatch.DeleteInsightRulesInput{
		RuleNames: []*string{f.name},
	})

	return err
}

func (f *CloudWatchInsightRule) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("Name", f.name)
	return properties
}

func (f *CloudWatchInsightRule) String() string {
	return *f.name
}
