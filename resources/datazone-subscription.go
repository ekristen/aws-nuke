package resources

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/datazone"
	"github.com/aws/aws-sdk-go-v2/service/datazone/types"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	libtypes "github.com/ekristen/libnuke/pkg/types"

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
			"DataZoneSubscriptionGrant", "DataZoneSubscriptionTarget",
		},
	})
}

type DataZoneSubscriptionLister struct{}

func (l *DataZoneSubscriptionLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := datazone.NewFromConfig(*opts.Config)
	resources := make([]resource.Resource, 0)

	// First, list all domains
	domainParams := &datazone.ListDomainsInput{
		MaxResults: aws.Int32(100),
	}

	domainPaginator := datazone.NewListDomainsPaginator(svc, domainParams)
	for domainPaginator.HasMorePages() {
		domainResp, err := domainPaginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		// For each domain, list subscriptions
		for _, domain := range domainResp.Items {
			subParams := &datazone.ListSubscriptionsInput{
				DomainIdentifier: domain.Id,
				MaxResults:       aws.Int32(100),
			}

			subPaginator := datazone.NewListSubscriptionsPaginator(svc, subParams)
			for subPaginator.HasMorePages() {
				subResp, err := subPaginator.NextPage(ctx)
				if err != nil {
					return nil, err // fail loudly for SCP denials
				}

				for _, sub := range subResp.Items {
					resources = append(resources, &DataZoneSubscription{
						svc:                 svc,
						DomainID:            domain.Id,
						ID:                  sub.Id,
						Status:              aws.String(string(sub.Status)),
						DomainName:          domain.Name,
						SubscribedPrincipal: sub.SubscribedPrincipal,
						SubscribedListing:   sub.SubscribedListing,
						CreatedAt:           sub.CreatedAt,
					})
				}
			}
		}
	}

	return resources, nil
}

type DataZoneSubscription struct {
	svc                 *datazone.Client
	DomainID            *string
	ID                  *string
	Status              *string
	DomainName          *string
	SubscribedPrincipal types.SubscribedPrincipal
	SubscribedListing   *types.SubscribedListing
	CreatedAt           *time.Time
}

func (r *DataZoneSubscription) Filter() error {
	//no pending or in-progress states for subscription, only cancelled, revoked or approved is available.
	if r.Status != nil && types.SubscriptionStatus(*r.Status) == types.SubscriptionStatusCancelled {
		return fmt.Errorf("subscription is already cancelled")
	}
	return nil
}

func (r *DataZoneSubscription) Remove(ctx context.Context) error {
	_, err := r.svc.CancelSubscription(ctx, &datazone.CancelSubscriptionInput{
		DomainIdentifier: r.DomainID,
		Identifier:       r.ID,
	})
	return err
}

func (r *DataZoneSubscription) Properties() libtypes.Properties {
	return libtypes.NewPropertiesFromStruct(r)
}
