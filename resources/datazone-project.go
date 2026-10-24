package resources

// This file teaches aws-nuke how to find and delete DataZone Projects.
//
// Projects live inside a Domain. To list projects, you first need the
// domain ID — that's why we call the shared listDataZoneDomainIDs helper
// defined in datazone-domain.go.

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

const DataZoneProjectResource = "DataZoneProject"

func init() {
	registry.Register(&registry.Registration{
		Name:     DataZoneProjectResource,
		Scope:    nuke.Account,
		Resource: &DataZoneProject{},
		Lister:   &DataZoneProjectLister{},
		// Projects have no children of their own that block deletion,
		// so no DependsOn needed here.
	})
}

type DataZoneProjectLister struct {
	mockSvc datazoneiface.DataZoneAPI
}

func (l *DataZoneProjectLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	var svc datazoneiface.DataZoneAPI
	if l.mockSvc != nil {
		svc = l.mockSvc
	} else {
		svc = datazone.New(opts.Session)
	}

	resources := make([]resource.Resource, 0)

	// First, get all domain IDs — we need to list projects per domain.
	domainIDs, err := listDataZoneDomainIDs(svc)
	if err != nil {
		return nil, err
	}

	// For each domain, list all projects inside it.
	for _, domainID := range domainIDs {
		params := &datazone.ListProjectsInput{
			DomainIdentifier: domainID,
			MaxResults:       aws.Int64(50),
		}

		for {
			resp, err := svc.ListProjects(params)
			if err != nil {
				return nil, err
			}

			for _, project := range resp.Items {
				resources = append(resources, &DataZoneProject{
					svc:      svc,
					DomainID: domainID,
					ID:       project.Id,
					Name:     project.Name,
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

// DataZoneProject represents a single DataZone project found in an account.
type DataZoneProject struct {
	svc      datazoneiface.DataZoneAPI
	DomainID *string
	ID       *string
	Name     *string
}

// Remove deletes this project. Both the domain ID and project ID are required.
func (r *DataZoneProject) Remove(_ context.Context) error {
	_, err := r.svc.DeleteProject(&datazone.DeleteProjectInput{
		DomainIdentifier: r.DomainID,
		Identifier:       r.ID,
	})
	return err
}

func (r *DataZoneProject) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *DataZoneProject) String() string {
	return aws.ToString(r.ID)
}
