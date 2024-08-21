package resources

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchevents"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const CloudWatchEventsRuleResource = "CloudWatchEventsRule"

func init() {
	registry.Register(&registry.Registration{
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
				svc:          svc,
				Name:         rule.Name,
				ARN:          rule.Arn,
				State:        rule.State,
				EventBusName: bus.Name,
			})
		}
	}
	return resources, nil
}

type CloudWatchEventsRule struct {
	svc          *cloudwatchevents.CloudWatchEvents
	Name         *string
	ARN          *string
	State        *string
	EventBusName *string
}

func (r *CloudWatchEventsRule) Remove(_ context.Context) error {
	_, err := r.svc.DeleteRule(&cloudwatchevents.DeleteRuleInput{
		Name:         r.Name,
		EventBusName: r.EventBusName,
		Force:        aws.Bool(true),
	})
	return err
}

func (r *CloudWatchEventsRule) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *CloudWatchEventsRule) String() string {
	// TODO: remove Rule:, mark as breaking change for filters
	return fmt.Sprintf("Rule: %s", *r.Name)
}
