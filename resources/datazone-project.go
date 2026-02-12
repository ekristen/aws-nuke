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

const DataZoneProjectResource = "DataZoneProject"

func init() {
	registry.Register(&registry.Registration{
		Name:     DataZoneProjectResource,
		Scope:    nuke.Account,
		Resource: &DataZoneProject{},
		Lister:   &DataZoneProjectLister{},
		DependsOn: []string{
			"DataZoneSubscription",
		},
	})
}

type DataZoneProjectLister struct{}

func (l *DataZoneProjectLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
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

		// For each domain, list its projects
		for _, domain := range domainResp.Items {
			projectParams := &datazone.ListProjectsInput{
				DomainIdentifier: domain.Id,
				MaxResults:       aws.Int64(100),
			}

			for {
				projectResp, err := svc.ListProjects(projectParams)
				if err != nil {
					return nil, err // Don't swallow errors - fail loudly for SCP denials
				}

				for _, project := range projectResp.Items {
					resources = append(resources, &DataZoneProject{
						svc:           svc,
						DomainID:      domain.Id,
						ID:            project.Id,
						Name:          project.Name,
						CreatedAt:     project.CreatedAt,
						CreatedBy:     project.CreatedBy,
						Description:   project.Description,
						DomainName:    domain.Name,
						ProjectStatus: project.ProjectStatus,
					})
				}

				if projectResp.NextToken == nil {
					break
				}

				projectParams.NextToken = projectResp.NextToken
			}
		}

		if domainResp.NextToken == nil {
			break
		}

		domainParams.NextToken = domainResp.NextToken
	}

	return resources, nil
}

type DataZoneProject struct {
	svc           *datazone.DataZone
	DomainID      *string
	ID            *string
	Name          *string
	CreatedAt     *time.Time
	CreatedBy     *string
	Description   *string
	DomainName    *string
	ProjectStatus *string
}

func (r *DataZoneProject) Remove(_ context.Context) error {
	_, err := r.svc.DeleteProject(&datazone.DeleteProjectInput{
		DomainIdentifier: r.DomainID,
		Identifier:       r.ID,
	})

	if err != nil {
		var awsErr awserr.Error
		if errors.As(err, &awsErr) {
			switch awsErr.Code() {
			case "ResourceNotFoundException":
				// Project already deleted
				return nil
			case "ConflictException":
				// Deletion may already be in progress or project has dependencies, accept it
				return nil
			}
		}
		return err
	}

	return nil
}


func (r *DataZoneProject) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("ID", r.ID)
	properties.Set("Name", r.Name)
	properties.Set("DomainID", r.DomainID)
	properties.Set("DomainName", r.DomainName)
	if r.ProjectStatus != nil {
		properties.Set("ProjectStatus", r.ProjectStatus)
	}
	if r.CreatedAt != nil {
		properties.Set("CreatedAt", r.CreatedAt.Format(time.RFC3339))
	}
	if r.CreatedBy != nil {
		properties.Set("CreatedBy", r.CreatedBy)
	}
	if r.Description != nil {
		properties.Set("Description", r.Description)
	}

	return properties
}
