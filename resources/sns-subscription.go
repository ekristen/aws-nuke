package resources

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/service/sns"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const SNSSubscriptionResource = "SNSSubscription"

func init() {
	registry.Register(&registry.Registration{
		Name:   SNSSubscriptionResource,
		Scope:  nuke.Account,
		Lister: &SNSSubscriptionLister{},
	})
}

type SNSSubscriptionLister struct{}

func (l *SNSSubscriptionLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := sns.New(opts.Session)

	params := &sns.ListSubscriptionsInput{}
	resources := make([]resource.Resource, 0)

	for {
		resp, err := svc.ListSubscriptions(params)
		if err != nil {
			return nil, err
		}

		for _, subscription := range resp.Subscriptions {
			if *subscription.SubscriptionArn != "PendingConfirmation" {
				resources = append(resources, &SNSSubscription{
					svc:      svc,
					ARN:      subscription.SubscriptionArn,
					Owner:    subscription.Owner,
					TopicARN: subscription.TopicArn,
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

type SNSSubscription struct {
	svc      *sns.SNS
	ARN      *string
	Owner    *string
	TopicARN *string
}

func (r *SNSSubscription) Remove(_ context.Context) error {
	_, err := r.svc.Unsubscribe(&sns.UnsubscribeInput{
		SubscriptionArn: r.ARN,
	})
	return err
}

func (r *SNSSubscription) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *SNSSubscription) String() string {
	return fmt.Sprintf("Owner: %s ARN: %s", *r.Owner, *r.ARN)
}
