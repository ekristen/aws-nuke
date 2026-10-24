package resources

// DataZone Subscriptions represent approved access requests to data assets.
// They must be cancelled before the domain can be deleted.

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/datazone"
	"github.com/aws/aws-sdk-go/service/datazone/datazoneiface"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const DataZoneSubscriptionResource = "DataZoneSubscription"

func init() {
	registry.Register(&registry.Registration{
		Name:     DataZoneSubscriptionResource,
		Scope:    nuke.Account,
		Resource: &DataZoneSubscription{},
		Lister:   &DataZoneSubscriptionLister{},
		DependsOn: []string{
			DataZoneSubscriptionGrantResource,
		},
	})
}

type DataZoneSubscriptionLister struct {
	mockSvc datazoneiface.DataZoneAPI
}

func (l *DataZoneSubscriptionLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	var svc datazoneiface.DataZoneAPI
	if l.mockSvc != nil {
		svc = l.mockSvc
	} else {
		svc = datazone.New(opts.Session)
	}

	resources := make([]resource.Resource, 0)

	domainIDs, err := listDataZoneDomainIDs(svc)
	if err != nil {
		return nil, err
	}

	for _, domainID := range domainIDs {
		params := &datazone.ListSubscriptionsInput{
			DomainIdentifier: domainID,
			MaxResults:       aws.Int64(50),
		}

		for {
			resp, err := svc.ListSubscriptions(params)
			if err != nil {
				return nil, err
			}

			for _, sub := range resp.Items {
				resources = append(resources, &DataZoneSubscription{
					svc:      svc,
					DomainID: domainID,
					ID:       sub.Id,
					Status:   aws.String(string(*sub.Status)),
				})
			}

			if resp.NextToken == nil {
				break
			}
			params.NextToken = resp.NextToken
		}
	}

	return resources, nil
}

type DataZoneSubscription struct {
	svc      datazoneiface.DataZoneAPI
	DomainID *string
	ID       *string
	Status   *string
}

// Remove cancels the subscription. DataZone uses CancelSubscription
// rather than DeleteSubscription for active subscriptions.
func (r *DataZoneSubscription) Remove(_ context.Context) error {
	_, err := r.svc.CancelSubscription(&datazone.CancelSubscriptionInput{
		DomainIdentifier: r.DomainID,
		Identifier:       r.ID,
	})
	return err
}

func (r *DataZoneSubscription) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *DataZoneSubscription) String() string {
	return aws.ToString(r.ID)
}
