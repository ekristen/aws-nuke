package resources

// This file teaches aws-nuke how to find and delete Amazon DataZone Domains.
//
// A DataZone Domain is the top-level container — it holds projects,
// assets, subscriptions etc. It must be deleted LAST (after projects
// and subscriptions are gone), which is why DependsOn lists the
// child resource types below.

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

// DataZoneDomainResource is the string name aws-nuke uses to identify
// this resource type in config files and logs.
const DataZoneDomainResource = "DataZoneDomain"

// init() runs automatically when the package loads. It registers this
// resource type with the aws-nuke registry so the tool knows it exists.
func init() {
	registry.Register(&registry.Registration{
		Name:     DataZoneDomainResource,
		Scope:    nuke.Account,
		Resource: &DataZoneDomain{},
		Lister:   &DataZoneDomainLister{},
		// DependsOn means: delete ALL of these resource types first,
		// before attempting to delete any DataZone domains.
		DependsOn: []string{
			DataZoneProjectResource,
			DataZoneSubscriptionGrantResource,
			DataZoneSubscriptionResource,
		},
	})
}

// DataZoneDomainLister is responsible for finding all DataZone domains
// in the AWS account.
type DataZoneDomainLister struct {
	mockSvc datazoneiface.DataZoneAPI
}

func (l *DataZoneDomainLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	var svc datazoneiface.DataZoneAPI
	if l.mockSvc != nil {
		svc = l.mockSvc
	} else {
		// Create a new DataZone client using the AWS session (credentials, region etc.)
		svc = datazone.New(opts.Session)
	}

	resources := make([]resource.Resource, 0)

	// ListDomains returns all DataZone domains in the account.
	// We page through results in case there are many domains.
	params := &datazone.ListDomainsInput{
		MaxResults: aws.Int64(50),
	}

	for {
		resp, err := svc.ListDomains(params)
		if err != nil {
			return nil, err
		}

		for _, domain := range resp.Items {
			resources = append(resources, &DataZoneDomain{
				svc:    svc,
				ID:     domain.Id,
				Name:   domain.Name,
				Status: aws.String(string(*domain.Status)),
			})
		}

		// If there's no NextToken, we've reached the last page.
		if resp.NextToken == nil {
			break
		}
		params.NextToken = resp.NextToken
	}

	return resources, nil
}

// DataZoneDomain represents a single DataZone domain that aws-nuke
// has found and will attempt to delete.
type DataZoneDomain struct {
	svc    datazoneiface.DataZoneAPI
	ID     *string
	Name   *string
	Status *string
}

// Remove calls the AWS API to delete this specific domain.
func (r *DataZoneDomain) Remove(_ context.Context) error {
	_, err := r.svc.DeleteDomain(&datazone.DeleteDomainInput{
		Identifier: r.ID,
	})
	return err
}

// Properties returns the fields aws-nuke displays and uses for filtering.
// For example, in your nuke config you could filter by Name to protect
// specific domains from deletion.
func (r *DataZoneDomain) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

// String is what gets printed in the aws-nuke output, e.g.:
//
//	DataZoneDomain - dzd_abc123 (my-domain)
func (r *DataZoneDomain) String() string {
	return aws.ToString(r.ID)
}

// listDataZoneDomainIDs is a shared helper used by child resource listers
// (projects, subscriptions etc.) that need to know all domain IDs first
// before they can list their own resources.
func listDataZoneDomainIDs(svc datazoneiface.DataZoneAPI) ([]*string, error) {
	var domainIDs []*string

	params := &datazone.ListDomainsInput{
		MaxResults: aws.Int64(50),
	}

	for {
		resp, err := svc.ListDomains(params)
		if err != nil {
			return nil, err
		}
		for _, domain := range resp.Items {
			domainIDs = append(domainIDs, domain.Id)
		}
		if resp.NextToken == nil {
			break
		}
		params.NextToken = resp.NextToken
	}

	return domainIDs, nil
}


