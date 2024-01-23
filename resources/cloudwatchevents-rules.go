package resources

import (
	"context"

	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchevents"

	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const CloudWatchEventsRuleResource = "CloudWatchEventsRule"

func init() {
	resource.Register(&resource.Registration{
		Name:   CloudWatchEventsRuleResource,
		Scope:  nuke.Account,
		Lister: &CloudWatchEventsRuleLister{},
	})
}

type CloudWatchEventsRuleLister struct{}

func (l *CloudWatchEventsRuleLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := cloudwatchevents.New(opts.Session)

	resp, err := svc.ListEventBuses(nil)
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, 0)
	for _, bus := range resp.EventBuses {
		resp, err := svc.ListRules(&cloudwatchevents.ListRulesInput{
			EventBusName: bus.Name,
		})
		if err != nil {
			return nil, err
		}

		for _, rule := range resp.Rules {
			resources = append(resources, &CloudWatchEventsRule{
				svc:     svc,
				name:    rule.Name,
				busName: bus.Name,
			})
		}
	}
	return resources, nil
}

type CloudWatchEventsRule struct {
	svc     *cloudwatchevents.CloudWatchEvents
	name    *string
	busName *string
}

func (rule *CloudWatchEventsRule) Remove(_ context.Context) error {
	_, err := rule.svc.DeleteRule(&cloudwatchevents.DeleteRuleInput{
		Name:         rule.name,
		EventBusName: rule.busName,
		Force:        aws.Bool(true),
	})
	return err
}

func (rule *CloudWatchEventsRule) String() string {
	// TODO: remove Rule:, mark as breaking change for filters
	return fmt.Sprintf("Rule: %s", *rule.name)
}
