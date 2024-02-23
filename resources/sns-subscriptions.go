package resources

import (
	"context"

	"fmt"

	"github.com/aws/aws-sdk-go/service/sns"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
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
					svc:  svc,
					id:   subscription.SubscriptionArn,
					name: subscription.Owner,
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
	svc  *sns.SNS
	id   *string
	name *string
}

func (subs *SNSSubscription) Remove(_ context.Context) error {
	_, err := subs.svc.Unsubscribe(&sns.UnsubscribeInput{
		SubscriptionArn: subs.id,
	})
	return err
}

func (subs *SNSSubscription) String() string {
	return fmt.Sprintf("Owner: %s ARN: %s", *subs.name, *subs.id)
}
