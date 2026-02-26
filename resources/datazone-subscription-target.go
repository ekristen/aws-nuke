package resources

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/datazone"
	"github.com/aws/aws-sdk-go-v2/service/datazone/types"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	libtypes "github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const DataZoneSubscriptionTargetResource = "DataZoneSubscriptionTarget"

func init() {
	registry.Register(&registry.Registration{
		Name:     DataZoneSubscriptionTargetResource,
		Scope:    nuke.Account,
		Resource: &DataZoneSubscriptionTarget{},
		Lister:   &DataZoneSubscriptionTargetLister{},
		DependsOn: []string{
			"DataZoneSubscriptionGrant",
		},
	})
}

type DataZoneSubscriptionTargetLister struct{}

func (l *DataZoneSubscriptionTargetLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
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

		// For each domain, collect all environments efficiently
		for di := range domainResp.Items {
			domain := &domainResp.Items[di]
			environments, err := l.listAllEnvironmentsInDomain(ctx, svc, domain.Id)
			if err != nil {
				return nil, err // Don't swallow errors - fail loudly for SCP denials
			}

			// Now list subscription targets for all environments
			for i := range environments {
				env := &environments[i]
				err := l.listSubscriptionTargetsInEnvironment(ctx, svc, domain, env, &resources)
				if err != nil {
					return nil, err // Don't swallow errors - fail loudly for SCP denials
				}
			}
		}
	}

	return resources, nil
}

// Helper function to list all environments in a domain across all projects
func (l *DataZoneSubscriptionTargetLister) listAllEnvironmentsInDomain(
	ctx context.Context, svc *datazone.Client, domainID *string,
) ([]types.EnvironmentSummary, error) {
	var allEnvironments []types.EnvironmentSummary

	// First, get all projects in the domain
	projectParams := &datazone.ListProjectsInput{
		DomainIdentifier: domainID,
		MaxResults:       aws.Int32(100),
	}

	projectPaginator := datazone.NewListProjectsPaginator(svc, projectParams)
	for projectPaginator.HasMorePages() {
		projectResp, err := projectPaginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		// For each project, get all environments
		for _, project := range projectResp.Items {
			envParams := &datazone.ListEnvironmentsInput{
				DomainIdentifier:  domainID,
				ProjectIdentifier: project.Id,
				MaxResults:        aws.Int32(100),
			}

			envPaginator := datazone.NewListEnvironmentsPaginator(svc, envParams)
			for envPaginator.HasMorePages() {
				envResp, err := envPaginator.NextPage(ctx)
				if err != nil {
					return nil, err
				}

				allEnvironments = append(allEnvironments, envResp.Items...)
			}
		}
	}

	return allEnvironments, nil
}

// Helper function to list subscription targets in a specific environment
func (l *DataZoneSubscriptionTargetLister) listSubscriptionTargetsInEnvironment(
	ctx context.Context,
	svc *datazone.Client,
	domain *types.DomainSummary,
	env *types.EnvironmentSummary,
	resources *[]resource.Resource,
) error {
	targetParams := &datazone.ListSubscriptionTargetsInput{
		DomainIdentifier:      domain.Id,
		EnvironmentIdentifier: env.Id,
		MaxResults:            aws.Int32(100),
	}

	targetPaginator := datazone.NewListSubscriptionTargetsPaginator(svc, targetParams)
	for targetPaginator.HasMorePages() {
		targetResp, err := targetPaginator.NextPage(ctx)
		if err != nil {
			return err
		}

		for i := range targetResp.Items {
			target := &targetResp.Items[i]
			*resources = append(*resources, &DataZoneSubscriptionTarget{
				svc:           svc,
				DomainID:      domain.Id,
				ID:            target.Id,
				Name:          target.Name,
				EnvironmentID: target.EnvironmentId,
				ProjectID:     target.ProjectId,
				DomainName:    domain.Name,
				Type:          target.Type,
				Provider:      target.Provider,
				CreatedAt:     target.CreatedAt,
			})
		}
	}

	return nil
}

type DataZoneSubscriptionTarget struct {
	svc           *datazone.Client
	DomainID      *string `property:"-"`
	EnvironmentID *string `property:"-"`
	ID            *string
	Name          *string
	ProjectID     *string
	DomainName    *string
	Type          *string
	Provider      *string
	CreatedAt     *time.Time
}

func (r *DataZoneSubscriptionTarget) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteSubscriptionTarget(ctx, &datazone.DeleteSubscriptionTargetInput{
		DomainIdentifier:      r.DomainID,
		EnvironmentIdentifier: r.EnvironmentID,
		Identifier:            r.ID,
	})
	return err
}

func (r *DataZoneSubscriptionTarget) Properties() libtypes.Properties {
	return libtypes.NewPropertiesFromStruct(r)
}
