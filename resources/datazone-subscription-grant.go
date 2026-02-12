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

const DataZoneSubscriptionGrantResource = "DataZoneSubscriptionGrant"

func init() {
	registry.Register(&registry.Registration{
		Name:     DataZoneSubscriptionGrantResource,
		Scope:    nuke.Account,
		Resource: &DataZoneSubscriptionGrant{},
		Lister:   &DataZoneSubscriptionGrantLister{},
	})
}

type DataZoneSubscriptionGrantLister struct{}

func (l *DataZoneSubscriptionGrantLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
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

		// For each domain, list subscription grants
		for _, domain := range domainResp.Items {
			grantParams := &datazone.ListSubscriptionGrantsInput{
				DomainIdentifier: domain.Id,
				MaxResults:       aws.Int64(100),
			}

			for {
				grantResp, err := svc.ListSubscriptionGrants(grantParams)
				if err != nil {
					return nil, err // Don't swallow errors - fail loudly for SCP denials
				}

				for _, grant := range grantResp.Items {
					resources = append(resources, &DataZoneSubscriptionGrant{
						svc:            svc,
						DomainID:       domain.Id,
						ID:             grant.Id,
						Status:         grant.Status,
						DomainName:     domain.Name,
						SubscriptionID: grant.SubscriptionId,
						GrantedEntity:  grant.GrantedEntity,
						CreatedAt:      grant.CreatedAt,
					})
				}

				if grantResp.NextToken == nil {
					break
				}

				grantParams.NextToken = grantResp.NextToken
			}
		}

		if domainResp.NextToken == nil {
			break
		}

		domainParams.NextToken = domainResp.NextToken
	}

	return resources, nil
}

type DataZoneSubscriptionGrant struct {
	svc            *datazone.DataZone
	DomainID       *string
	ID             *string
	Status         *string
	DomainName     *string
	SubscriptionID *string
	GrantedEntity  *datazone.GrantedEntity
	CreatedAt      *time.Time
}

func (r *DataZoneSubscriptionGrant) Remove(_ context.Context) error {
	_, err := r.svc.DeleteSubscriptionGrant(&datazone.DeleteSubscriptionGrantInput{
		DomainIdentifier: r.DomainID,
		Identifier:       r.ID,
	})

	if err != nil {
		var awsErr awserr.Error
		if errors.As(err, &awsErr) {
			switch awsErr.Code() {
			case "ResourceNotFoundException":
				// Grant already deleted
				return nil
			case "ConflictException":
				// Deletion may already be in progress, accept it
				return nil
			}
		}
		return err
	}

	return nil
}


func (r *DataZoneSubscriptionGrant) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("ID", r.ID)
	properties.Set("DomainID", r.DomainID)
	properties.Set("DomainName", r.DomainName)
	if r.Status != nil {
		properties.Set("Status", r.Status)
	}
	if r.SubscriptionID != nil {
		properties.Set("SubscriptionID", r.SubscriptionID)
	}
	if r.CreatedAt != nil {
		properties.Set("CreatedAt", r.CreatedAt.Format(time.RFC3339))
	}

	return properties
}
