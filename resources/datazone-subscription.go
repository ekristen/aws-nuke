package resources

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go/aws"              //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/datazone" //nolint:staticcheck
	"github.com/aws/aws-sdk-go/aws/awserr"       //nolint:staticcheck

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
			"DataZoneSubscriptionGrant", "DataZoneSubscriptionTarget",
		},
	})
}

type DataZoneSubscriptionLister struct{}

func (l *DataZoneSubscriptionLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := datazone.New(opts.Session)
	resources := make([]resource.Resource, 0)

	// First, list all domains
	domainParams := &datazone.ListDomainsInput{
		MaxResults: aws.Int64(100),
	}

	for {
		domainResp, err := svc.ListDomains(domainParams)
		if err != nil {
			return nil, err
		}

		// For each domain, list subscriptions
		for _, domain := range domainResp.Items {
			subParams := &datazone.ListSubscriptionsInput{
				DomainIdentifier: domain.Id,
				MaxResults:       aws.Int64(100),
			}

			for {
				subResp, err := svc.ListSubscriptions(subParams)
				if err != nil {
					return nil, err // Don't swallow errors - fail loudly for SCP denials
				}

				for _, sub := range subResp.Items {
					resources = append(resources, &DataZoneSubscription{
						svc:                 svc,
						DomainID:            domain.Id,
						ID:                  sub.Id,
						Status:              sub.Status,
						DomainName:          domain.Name,
						SubscribedPrincipal: sub.SubscribedPrincipal,
						SubscribedListing:   sub.SubscribedListing,
						CreatedAt:           sub.CreatedAt,
					})
				}

				if subResp.NextToken == nil {
					break
				}

				subParams.NextToken = subResp.NextToken
			}
		}

		if domainResp.NextToken == nil {
			break
		}

		domainParams.NextToken = domainResp.NextToken
	}

	return resources, nil
}

type DataZoneSubscription struct {
	svc                 *datazone.DataZone
	DomainID            *string
	ID                  *string
	Status              *string
	DomainName          *string
	SubscribedPrincipal *datazone.SubscribedPrincipal
	SubscribedListing   *datazone.SubscribedListing
	CreatedAt           *time.Time
}

func (r *DataZoneSubscription) Remove(_ context.Context) error {
	// Only skip if already in a terminal cancelled state
	if r.Status != nil && *r.Status == "CANCELLED" {
		return nil
	}

	_, err := r.svc.CancelSubscription(&datazone.CancelSubscriptionInput{
		DomainIdentifier: r.DomainID,
		Identifier:       r.ID,
	})

	if err != nil {
		var awsErr awserr.Error
		if errors.As(err, &awsErr) {
			switch awsErr.Code() {
			case "ResourceNotFoundException":
				// Subscription already deleted
				return nil
			case "ConflictException":
				// Cancellation may already be in progress, accept it
				return nil
			}
		}
		return err
	}

	// Note: CancelSubscription is async - subscription will transition to CANCELLED status
	// aws-nuke will re-run and pick up the final state in subsequent passes
	return nil
}


func (r *DataZoneSubscription) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("ID", r.ID)
	properties.Set("DomainID", r.DomainID)
	properties.Set("DomainName", r.DomainName)
	if r.Status != nil {
		properties.Set("Status", r.Status)
	}
	if r.CreatedAt != nil {
		properties.Set("CreatedAt", r.CreatedAt.Format(time.RFC3339))
	}

	return properties
}
