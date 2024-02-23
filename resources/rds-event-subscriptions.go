package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const RDSEventSubscriptionResource = "RDSEventSubscription"

func init() {
	registry.Register(&registry.Registration{
		Name:   RDSEventSubscriptionResource,
		Scope:  nuke.Account,
		Lister: &RDSEventSubscriptionLister{},
	})
}

type RDSEventSubscriptionLister struct{}

type RDSEventSubscription struct {
	svc     *rds.RDS
	id      *string
	enabled *bool
	tags    []*rds.Tag
}

func (l *RDSEventSubscriptionLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := rds.New(opts.Session)

	params := &rds.DescribeEventSubscriptionsInput{
		MaxRecords: aws.Int64(100),
	}
	resp, err := svc.DescribeEventSubscriptions(params)
	if err != nil {
		return nil, err
	}

	var resources []resource.Resource
	for _, eventSubscription := range resp.EventSubscriptionsList {
		tags, err := svc.ListTagsForResource(&rds.ListTagsForResourceInput{
			ResourceName: eventSubscription.EventSubscriptionArn,
		})
		if err != nil {
			continue
		}

		resources = append(resources, &RDSEventSubscription{
			svc:     svc,
			id:      eventSubscription.CustSubscriptionId,
			enabled: eventSubscription.Enabled,
			tags:    tags.TagList,
		})
	}

	return resources, nil
}

func (i *RDSEventSubscription) Remove(_ context.Context) error {
	params := &rds.DeleteEventSubscriptionInput{
		SubscriptionName: i.id,
	}

	_, err := i.svc.DeleteEventSubscription(params)
	if err != nil {
		return err
	}

	return nil
}

func (i *RDSEventSubscription) String() string {
	return *i.id
}

func (i *RDSEventSubscription) Properties() types.Properties {
	properties := types.NewProperties()
	properties.
		Set("ID", i.id).
		Set("Enabled", i.enabled)

	for _, tag := range i.tags {
		properties.SetTag(tag.Key, tag.Value)
	}

	return properties
}
