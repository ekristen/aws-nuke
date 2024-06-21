package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iot"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const IoTTopicRuleResource = "IoTTopicRule"

func init() {
	registry.Register(&registry.Registration{
		Name:   IoTTopicRuleResource,
		Scope:  nuke.Account,
		Lister: &IoTTopicRuleLister{},
	})
}

type IoTTopicRuleLister struct{}

func (l *IoTTopicRuleLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := iot.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &iot.ListTopicRulesInput{
		MaxResults: aws.Int64(100),
	}
	for {
		output, err := svc.ListTopicRules(params)
		if err != nil {
			return nil, err
		}

		for _, rule := range output.Rules {
			resources = append(resources, &IoTTopicRule{
				svc:  svc,
				name: rule.RuleName,
			})
		}
		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type IoTTopicRule struct {
	svc  *iot.IoT
	name *string
}

func (f *IoTTopicRule) Remove(_ context.Context) error {
	_, err := f.svc.DeleteTopicRule(&iot.DeleteTopicRuleInput{
		RuleName: f.name,
	})

	return err
}

func (f *IoTTopicRule) String() string {
	return *f.name
}
