package resources

import (
	"context"
	"errors"
	"fmt"

	"github.com/gotidy/ptr"

	"github.com/aws/aws-sdk-go/aws"                      //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/cloudwatchevents" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const CloudWatchEventsRuleResource = "CloudWatchEventsRule"

func init() {
	registry.Register(&registry.Registration{
		Name:     CloudWatchEventsRuleResource,
		Scope:    nuke.Account,
		Resource: &CloudWatchEventsRule{},
		Lister:   &CloudWatchEventsRuleLister{},
	})
}

type CloudWatchEventsRuleLister struct{}

func (l *CloudWatchEventsRuleLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	var resources []resource.Resource

	svc := cloudwatchevents.New(opts.Session)

	params := &cloudwatchevents.ListEventBusesInput{}

	for {
		resp, err := svc.ListEventBuses(params)
		if err != nil {
			return nil, err
		}

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
					ManagedBy:    rule.ManagedBy,
				})
			}
		}

		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return resources, nil
}

type CloudWatchEventsRule struct {
	svc          *cloudwatchevents.CloudWatchEvents
	Name         *string `description:"The name of the rule"`
	ARN          *string `description:"The ARN of the rule"`
	State        *string `description:"The state of the rule, e.g. ENABLED or DISABLED"`
	EventBusName *string `description:"The name of the event bus the rule belongs to"`
	ManagedBy    *string `description:"The principal of the AWS service that created and manages the rule, if any"`
}

func (r *CloudWatchEventsRule) Filter() error {
	if ptr.ToString(r.ManagedBy) != "" {
		return errors.New("cannot delete rule managed by " + ptr.ToString(r.ManagedBy))
	}
	return nil
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
