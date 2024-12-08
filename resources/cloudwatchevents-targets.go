package resources

import (
	"context"

	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchevents"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const CloudWatchEventsTargetResource = "CloudWatchEventsTarget"

func init() {
	registry.Register(&registry.Registration{
		Name:     CloudWatchEventsTargetResource,
		Scope:    nuke.Account,
		Resource: &CloudWatchEventsTarget{},
		Lister:   &CloudWatchEventsTargetLister{},
	})
}

type CloudWatchEventsTargetLister struct{}

func (l *CloudWatchEventsTargetLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
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
			targetResp, err := svc.ListTargetsByRule(&cloudwatchevents.ListTargetsByRuleInput{
				Rule:         rule.Name,
				EventBusName: bus.Name,
			})
			if err != nil {
				return nil, err
			}
			for _, target := range targetResp.Targets {
				resources = append(resources, &CloudWatchEventsTarget{
					svc:      svc,
					ruleName: rule.Name,
					targetID: target.Id,
					busName:  bus.Name,
				})
			}
		}
	}

	return resources, nil
}

type CloudWatchEventsTarget struct {
	svc      *cloudwatchevents.CloudWatchEvents
	targetID *string
	ruleName *string
	busName  *string
}

func (target *CloudWatchEventsTarget) Remove(_ context.Context) error {
	ids := []*string{target.targetID}
	_, err := target.svc.RemoveTargets(&cloudwatchevents.RemoveTargetsInput{
		Ids:          ids,
		Rule:         target.ruleName,
		EventBusName: target.busName,
		Force:        aws.Bool(true),
	})
	return err
}

func (target *CloudWatchEventsTarget) String() string {
	// TODO: change this to IAM format rule -> target and mark as breaking change for filters
	// TODO: add properties for rule and target separately
	return fmt.Sprintf("Rule: %s Target ID: %s", *target.ruleName, *target.targetID)
}
