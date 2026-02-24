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

func (l *DataZoneDomainLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := datazone.NewFromConfig(*opts.Config)
	resources := make([]resource.Resource, 0)

	params := &datazone.ListDomainsInput{
		MaxResults: aws.Int32(100),
	}

	paginator := datazone.NewListDomainsPaginator(svc, params)
	for paginator.HasMorePages() {
		resp, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, domain := range resp.Items {
			resources = append(resources, &DataZoneDomain{
				svc:         svc,
				ID:          domain.Id,
				Name:        domain.Name,
				Status:      aws.String(string(domain.Status)),
				CreatedAt:   domain.CreatedAt,
				Description: domain.Description,
			})
		}
	}

	return resources, nil
}

type DataZoneDomain struct {
	svc         *datazone.Client
	ID          *string
	Name        *string
	Status      *string
	CreatedAt   *time.Time
	Description *string
}

func (r *DataZoneDomain) Filter() error {
	if r.Status != nil {
		switch types.DomainStatus(*r.Status) {
		//domains in inprogress states are handled in HandleWait, so we want to skip them here to avoid false positives
		case types.DomainStatusDeleted:
			return fmt.Errorf("domain is already deleted")
		}
	}
	return nil
}

func (r *DataZoneDomain) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteDomain(ctx, &datazone.DeleteDomainInput{
		Identifier: r.ID,
	})
	return err
}

func (r *DataZoneDomain) HandleWait(ctx context.Context) error {
	resp, err := r.svc.GetDomain(ctx, &datazone.GetDomainInput{
		Identifier: r.ID,
	})
	if err != nil {
		return err
	}

	r.Status = aws.String(string(resp.Status))

	switch resp.Status {
	case types.DomainStatusDeleted:
		return nil
	case types.DomainStatusDeleting:
		return liberror.ErrWaitResource("domain deletion in progress")
	default:
		return liberror.ErrWaitResource(fmt.Sprintf("domain status: %s", resp.Status))
	}
}

func (r *DataZoneDomain) Properties() libtypes.Properties {
	return libtypes.NewPropertiesFromStruct(r)
}
