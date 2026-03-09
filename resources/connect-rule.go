package resources

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/connect"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const ConnectRuleResource = "ConnectRule"

func init() {
	registry.Register(&registry.Registration{
		Name:     ConnectRuleResource,
		Scope:    nuke.Account,
		Resource: &ConnectRule{},
		Lister:   &ConnectRuleLister{},
	})
}

type ConnectRuleLister struct{}

func (l *ConnectRuleLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := connect.NewFromConfig(*opts.Config)
	var resources []resource.Resource

	instances, err := listConnectInstances(ctx, svc)
	if err != nil {
		return nil, err
	}

	for _, instance := range instances {
		paginator := connect.NewListRulesPaginator(svc, &connect.ListRulesInput{
			InstanceId: instance.Id,
		})

		for paginator.HasMorePages() {
			resp, err := paginator.NextPage(ctx)
			if err != nil {
				return nil, err
			}

			for _, rule := range resp.RuleSummaryList {
				resources = append(resources, &ConnectRule{
					svc:             svc,
					InstanceID:      instance.Id,
					RuleID:          rule.RuleId,
					Name:            rule.Name,
					PublishStatus:   string(rule.PublishStatus),
					EventSourceName: string(rule.EventSourceName),
					CreatedAt:       rule.CreatedTime,
					UpdatedAt:       rule.LastUpdatedTime,
				})
			}
		}
	}

	return resources, nil
}

type ConnectRule struct {
	svc             *connect.Client
	InstanceID      *string
	RuleID          *string
	Name            *string
	PublishStatus   string
	EventSourceName string
	CreatedAt       *time.Time
	UpdatedAt       *time.Time
}

func (r *ConnectRule) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteRule(ctx, &connect.DeleteRuleInput{
		InstanceId: r.InstanceID,
		RuleId:     r.RuleID,
	})
	return err
}

func (r *ConnectRule) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *ConnectRule) String() string {
	return *r.Name
}
