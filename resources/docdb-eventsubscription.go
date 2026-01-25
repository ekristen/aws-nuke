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
			tagList := DocDBEmptyTags
			tags, err := svc.ListTagsForResource(ctx, &docdb.ListTagsForResourceInput{
				ResourceName: subscription.CustSubscriptionId,
			})
			if err == nil {
				tagList = tags.TagList
			}
			resources = append(resources, &DocDBEventSubscription{
				svc:             svc,
				ARN:             subscription.EventSubscriptionArn,
				Name:            subscription.CustSubscriptionId,
				SnsTopicArn:     subscription.SnsTopicArn,
				SourceType:      subscription.SourceType,
				Status:          subscription.Status,
				EventCategories: subscription.EventCategoriesList,
				Enabled:         subscription.Enabled,
				Tags:            tagList,
			})
		}
	}
	return resources, nil
}

type DocDBEventSubscription struct {
	svc *docdb.Client

	ARN             *string
	Name            *string
	SnsTopicArn     *string
	SourceType      *string
	Status          *string
	EventCategories []string
	Enabled         *bool
	Tags            []docdbtypes.Tag
}

func (r *DocDBEventSubscription) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteEventSubscription(ctx, &docdb.DeleteEventSubscriptionInput{
		SubscriptionName: r.Name,
	})
	return err
}

func (r *DocDBEventSubscription) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *DocDBEventSubscription) String() string {
	return *r.Name
}
