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

const DataZoneDomainResource = "DataZoneDomain"

func init() {
	registry.Register(&registry.Registration{
		Name:     DataZoneDomainResource,
		Scope:    nuke.Account,
		Resource: &DataZoneDomain{},
		Lister:   &DataZoneDomainLister{},
		DependsOn: []string{
			"DataZoneProject", "DataZoneSubscription",
		},
	})
}

type DataZoneDomainLister struct{}

func (l *DataZoneDomainLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := datazone.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &datazone.ListDomainsInput{
		MaxResults: aws.Int64(100),
	}

	for {
		resp, err := svc.ListDomains(params)
		if err != nil {
			return nil, err
		}

		for _, domain := range resp.Items {
			resources = append(resources, &DataZoneDomain{
				svc:         svc,
				ID:          domain.Id,
				Name:        domain.Name,
				Status:      domain.Status,
				CreatedAt:   domain.CreatedAt,
				Description: domain.Description,
			})
		}

		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return resources, nil
}

type DataZoneDomain struct {
	svc         *datazone.DataZone
	ID          *string
	Name        *string
	Status      *string
	CreatedAt   *time.Time
	Description *string
}

func (r *DataZoneDomain) Remove(_ context.Context) error {
	_, err := r.svc.DeleteDomain(&datazone.DeleteDomainInput{
		Identifier: r.ID,
	})

	if err != nil {
		// Handle AWS errors - check if deletion is already in progress or complete
		var awsErr awserr.Error
		if errors.As(err, &awsErr) {
			// If domain is already being deleted or doesn't exist, don't fail
			switch awsErr.Code() {
			case "ResourceNotFoundException":
				// Domain already deleted
				return nil
			case "ConflictException":
				// Check if it's already being deleted
				if r.Status != nil && *r.Status == "DELETING" {
					// Deletion in progress, let it complete
					return nil
				}
			}
		}
		return err
	}

	// Note: DeleteDomain is async - domain will transition through states
	// aws-nuke will re-run and pick up the final state in subsequent passes
	return nil
}

func (r *DataZoneDomain) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("ID", r.ID)
	properties.Set("Name", r.Name)
	properties.Set("Status", r.Status)
	if r.CreatedAt != nil {
		properties.Set("CreatedAt", r.CreatedAt.Format(time.RFC3339))
	}
	if r.Description != nil {
		properties.Set("Description", r.Description)
	}

	return properties
}

