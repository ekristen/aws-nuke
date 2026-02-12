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

func (l *DataZoneSubscriptionTargetLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
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

		// For each domain, collect all environments efficiently
		for _, domain := range domainResp.Items {
			environments, err := l.listAllEnvironmentsInDomain(svc, domain.Id)
			if err != nil {
				return nil, err // Don't swallow errors - fail loudly for SCP denials
			}

			// Now list subscription targets for all environments
			for _, env := range environments {
				err := l.listSubscriptionTargetsInEnvironment(svc, domain, env, &resources)
				if err != nil {
					return nil, err // Don't swallow errors - fail loudly for SCP denials
				}
			}
		}

		if domainResp.NextToken == nil {
			break
		}

		domainParams.NextToken = domainResp.NextToken
	}

	return resources, nil
}

// Helper function to list all environments in a domain across all projects
func (l *DataZoneSubscriptionTargetLister) listAllEnvironmentsInDomain(svc *datazone.DataZone, domainID *string) ([]*datazone.EnvironmentSummary, error) {
	var allEnvironments []*datazone.EnvironmentSummary

	// First, get all projects in the domain
	projectParams := &datazone.ListProjectsInput{
		DomainIdentifier: domainID,
		MaxResults:       aws.Int64(100),
	}

	for {
		projectResp, err := svc.ListProjects(projectParams)
		if err != nil {
			return nil, err
		}

		// For each project, get all environments
		for _, project := range projectResp.Items {
			envParams := &datazone.ListEnvironmentsInput{
				DomainIdentifier:  domainID,
				ProjectIdentifier: project.Id,
				MaxResults:        aws.Int64(100),
			}

			for {
				envResp, err := svc.ListEnvironments(envParams)
				if err != nil {
					return nil, err
				}

				allEnvironments = append(allEnvironments, envResp.Items...)

				if envResp.NextToken == nil {
					break
				}

				envParams.NextToken = envResp.NextToken
			}
		}

		if projectResp.NextToken == nil {
			break
		}

		projectParams.NextToken = projectResp.NextToken
	}

	return allEnvironments, nil
}

// Helper function to list subscription targets in a specific environment
func (l *DataZoneSubscriptionTargetLister) listSubscriptionTargetsInEnvironment(
	svc *datazone.DataZone,
	domain *datazone.DomainSummary,
	env *datazone.EnvironmentSummary,
	resources *[]resource.Resource,
) error {
	targetParams := &datazone.ListSubscriptionTargetsInput{
		DomainIdentifier:      domain.Id,
		EnvironmentIdentifier: env.Id,
		MaxResults:            aws.Int64(100),
	}

	for {
		targetResp, err := svc.ListSubscriptionTargets(targetParams)
		if err != nil {
			return err
		}

		for _, target := range targetResp.Items {
			*resources = append(*resources, &DataZoneSubscriptionTarget{
				svc:           svc,
				DomainID:      domain.Id,
				ID:            target.Id,
				Name:          target.Name,
				EnvironmentID: env.Id,
				ProjectID:     env.ProjectId,
				DomainName:    domain.Name,
				Type:          target.Type,
				Provider:      target.Provider,
				CreatedAt:     target.CreatedAt,
			})
		}

		if targetResp.NextToken == nil {
			break
		}

		targetParams.NextToken = targetResp.NextToken
	}

	return nil
}

type DataZoneSubscriptionTarget struct {
	svc           *datazone.DataZone
	DomainID      *string
	ID            *string
	Name          *string
	EnvironmentID *string
	ProjectID     *string
	DomainName    *string
	Type          *string
	Provider      *string
	CreatedAt     *time.Time
}

func (r *DataZoneSubscriptionTarget) Remove(_ context.Context) error {
	_, err := r.svc.DeleteSubscriptionTarget(&datazone.DeleteSubscriptionTargetInput{
		DomainIdentifier:      r.DomainID,
		EnvironmentIdentifier: r.EnvironmentID,
		Identifier:            r.ID,
	})

	if err != nil {
		var awsErr awserr.Error
		if errors.As(err, &awsErr) {
			switch awsErr.Code() {
			case "ResourceNotFoundException":
				// Target already deleted
				return nil
			case "ConflictException":
				// Deletion may already be in progress or target has dependencies, accept it
				return nil
			}
		}
		return err
	}

	return nil
}


func (r *DataZoneSubscriptionTarget) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("ID", r.ID)
	properties.Set("Name", r.Name)
	properties.Set("DomainID", r.DomainID)
	properties.Set("DomainName", r.DomainName)
	properties.Set("EnvironmentID", r.EnvironmentID)
	properties.Set("ProjectID", r.ProjectID)
	if r.Type != nil {
		properties.Set("Type", r.Type)
	}
	if r.Provider != nil {
		properties.Set("Provider", r.Provider)
	}
	if r.CreatedAt != nil {
		properties.Set("CreatedAt", r.CreatedAt.Format(time.RFC3339))
	}

	return properties
}
