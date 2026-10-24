package resources

// DataZone Subscription Grants represent the actual data access grants
// created when a subscription is approved. Must be deleted before subscriptions.

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

const DataZoneSubscriptionGrantResource = "DataZoneSubscriptionGrant"

func init() {
	registry.Register(&registry.Registration{
		Name:     DataZoneSubscriptionGrantResource,
		Scope:    nuke.Account,
		Resource: &DataZoneSubscriptionGrant{},
		Lister:   &DataZoneSubscriptionGrantLister{},
	})
}

type DataZoneSubscriptionGrantLister struct {
	mockSvc datazoneiface.DataZoneAPI
}

func (l *DataZoneSubscriptionGrantLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	var svc datazoneiface.DataZoneAPI
	if l.mockSvc != nil {
		svc = l.mockSvc
	} else {
		svc = datazone.New(opts.Session)
	}

	subscriptions, err := listDataZoneSubscriptions(svc)
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, 0)
	seen := make(map[string]struct{})
	for _, sub := range subscriptions {
		params := &datazone.ListSubscriptionGrantsInput{
			DomainIdentifier: sub.DomainID,
			SubscriptionId:   sub.ID,
			MaxResults:       aws.Int64(25),
		}

		for {
			resp, err := svc.ListSubscriptionGrants(params)
			if err != nil {
				return nil, err
			}

			for _, grant := range resp.Items {
				key := aws.StringValue(sub.DomainID) + "/" + aws.StringValue(grant.Id)
				if _, ok := seen[key]; ok {
					continue
				}
				seen[key] = struct{}{}
				resources = append(resources, &DataZoneSubscriptionGrant{
					svc:      svc,
					DomainID: sub.DomainID,
					ID:       grant.Id,
					Status:   aws.String(aws.StringValue(grant.Status)),
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

type DataZoneSubscriptionGrant struct {
	svc      datazoneiface.DataZoneAPI
	DomainID *string
	ID       *string
	Status   *string
}

func (r *DataZoneSubscriptionGrant) Remove(_ context.Context) error {
	_, err := r.svc.DeleteSubscriptionGrant(&datazone.DeleteSubscriptionGrantInput{
		DomainIdentifier: r.DomainID,
		Identifier:       r.ID,
	})
	return err
}

func (r *DataZoneSubscriptionGrant) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *DataZoneSubscriptionGrant) String() string {
	return aws.StringValue(r.ID)
}
