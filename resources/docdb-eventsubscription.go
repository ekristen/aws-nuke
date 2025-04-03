package resources

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/docdb"
	docdbtypes "github.com/aws/aws-sdk-go-v2/service/docdb/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"
)

const DocDBEventSubscriptionResource = "DocDBEventSubscription"

func init() {
	registry.Register(&registry.Registration{
		Name:     DocDBEventSubscriptionResource,
		Scope:    nuke.Account,
		Resource: &DocDBEventSubscription{},
		Lister:   &DocDBEventSubscriptionLister{},
	})
}

type DocDBEventSubscriptionLister struct{}

func (l *DocDBEventSubscriptionLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := docdb.NewFromConfig(*opts.Config)
	var resources []resource.Resource

	paginator := docdb.NewDescribeEventSubscriptionsPaginator(svc, &docdb.DescribeEventSubscriptionsInput{})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, subscription := range page.EventSubscriptionsList {
			tags, err := svc.ListTagsForResource(ctx, &docdb.ListTagsForResourceInput{
				ResourceName: subscription.CustSubscriptionId,
			})
			if err != nil {
				continue
			}
			resources = append(resources, &DocDBEventSubscription{
				svc:          svc,
				subscription: subscription,
				tags:         tags.TagList,
			})
		}
	}
	return resources, nil
}

type DocDBEventSubscription struct {
	svc          *docdb.Client
	subscription docdbtypes.EventSubscription
	tags         []docdbtypes.Tag
}

func (r *DocDBEventSubscription) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteEventSubscription(ctx, &docdb.DeleteEventSubscriptionInput{
		SubscriptionName: r.subscription.CustSubscriptionId,
	})
	return err
}

func (r *DocDBEventSubscription) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("ARN", r.subscription.EventSubscriptionArn)
	properties.Set("Name", r.subscription.CustSubscriptionId)
	properties.Set("SnsTopicArn", r.subscription.SnsTopicArn)
	properties.Set("SourceType", r.subscription.SourceType)
	properties.Set("Status", r.subscription.Status)
	properties.Set("EventCategories", r.subscription.EventCategoriesList)
	properties.Set("Enabled", r.subscription.Enabled)

	for _, tag := range r.tags {
		properties.SetTag(tag.Key, tag.Value)
	}

	return properties
}
