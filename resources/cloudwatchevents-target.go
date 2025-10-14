package resources

import (
	"context"
	"fmt"

	"github.com/gotidy/ptr"

	"github.com/aws/aws-sdk-go/service/cloudwatchevents" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

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
					Name:     rule.Name,
					TargetID: target.Id,
					BusName:  bus.Name,
				})
			}
		}
	}

	return resources, nil
}

type CloudWatchEventsTarget struct {
	svc      *cloudwatchevents.CloudWatchEvents
	TargetID *string `description:"The ID of the target for the rule"`
	Name     *string `description:"The name of the rule"`
	BusName  *string `description:"The name of the event bus the rule applies to"`
}

func (r *CloudWatchEventsTarget) Remove(_ context.Context) error {
	ids := []*string{r.TargetID}
	_, err := r.svc.RemoveTargets(&cloudwatchevents.RemoveTargetsInput{
		Ids:          ids,
		Rule:         r.Name,
		EventBusName: r.BusName,
		Force:        ptr.Bool(true),
	})
	return err
}

func (r *CloudWatchEventsTarget) String() string {
	// TODO: change this to IAM format rule -> target and mark as breaking change for filters
	return fmt.Sprintf("Rule: %s Target ID: %s", *r.Name, *r.TargetID)
}

func (r *CloudWatchEventsTarget) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}
