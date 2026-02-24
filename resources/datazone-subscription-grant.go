package resources

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/datazone"
	"github.com/aws/aws-sdk-go-v2/service/datazone/types"

	liberror "github.com/ekristen/libnuke/pkg/errors"
	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	libtypes "github.com/ekristen/libnuke/pkg/types"

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

func (l *DataZoneSubscriptionGrantLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
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

		// For each domain, list subscription grants
		for _, domain := range domainResp.Items {
			grantParams := &datazone.ListSubscriptionGrantsInput{
				DomainIdentifier: domain.Id,
				MaxResults:       aws.Int32(100),
			}

			grantPaginator := datazone.NewListSubscriptionGrantsPaginator(svc, grantParams)
			for grantPaginator.HasMorePages() {
				grantResp, err := grantPaginator.NextPage(ctx)
				if err != nil {
					return nil, err // Don't swallow errors - fail loudly for SCP denials
				}

				for _, grant := range grantResp.Items {
					resources = append(resources, &DataZoneSubscriptionGrant{
						svc:            svc,
						DomainID:       domain.Id,
						ID:             grant.Id,
						Status:         aws.String(string(grant.Status)),
						DomainName:     domain.Name,
						SubscriptionID: grant.SubscriptionId,
						GrantedEntity:  grant.GrantedEntity,
						CreatedAt:      grant.CreatedAt,
					})
				}
			}
		}
	}

	return resources, nil
}

type DataZoneSubscriptionGrant struct {
	svc            *datazone.Client
	DomainID       *string
	ID             *string
	Status         *string
	DomainName     *string
	SubscriptionID *string
	GrantedEntity  types.GrantedEntity
	CreatedAt      *time.Time
}

func (r *DataZoneSubscriptionGrant) Filter() error {
	if r.Status != nil {
		switch types.SubscriptionGrantOverallStatus(*r.Status) {
		case types.SubscriptionGrantOverallStatusInaccessible:
			return fmt.Errorf("subscription grant is inaccessible")
		case types.SubscriptionGrantOverallStatusGrantFailed:
			return fmt.Errorf("subscription grant is in grant failed state")
		case types.SubscriptionGrantOverallStatusRevokeFailed:
			return fmt.Errorf("subscription grant is in revoke failed state")
		case types.SubscriptionGrantOverallStatusGrantAndRevokeFailed:
			return fmt.Errorf("subscription grant is in grant and revoke failed state")
		}
	}
	return nil
}

func (r *DataZoneSubscriptionGrant) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteSubscriptionGrant(ctx, &datazone.DeleteSubscriptionGrantInput{
		DomainIdentifier: r.DomainID,
		Identifier:       r.ID,
	})
	return err
}

func (r *DataZoneSubscriptionGrant) HandleWait(ctx context.Context) error {
	resp, err := r.svc.GetSubscriptionGrant(ctx, &datazone.GetSubscriptionGrantInput{
		DomainIdentifier: r.DomainID,
		Identifier:       r.ID,
	})
	if err != nil {
		return err
	}

	r.Status = aws.String(string(resp.Status))

	switch resp.Status {
	case types.SubscriptionGrantOverallStatusGrantFailed,
		types.SubscriptionGrantOverallStatusRevokeFailed,
		types.SubscriptionGrantOverallStatusGrantAndRevokeFailed:
		return fmt.Errorf("subscription grant deletion failed (status=%s)", resp.Status)

	case types.SubscriptionGrantOverallStatusPending,
		types.SubscriptionGrantOverallStatusInProgress:
		return liberror.ErrWaitResource(fmt.Sprintf("subscription grant status=%s", resp.Status))

	default:
		// Still exists but in some other state, keep waiting
		return liberror.ErrWaitResource(fmt.Sprintf("subscription grant status=%s", resp.Status))
	}
}

func (r *DataZoneSubscriptionGrant) Properties() libtypes.Properties {
	return libtypes.NewPropertiesFromStruct(r)
}
