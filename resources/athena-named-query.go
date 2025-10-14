package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/athena" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const AthenaNamedQueryResource = "AthenaNamedQuery"

func init() {
	registry.Register(&registry.Registration{
		Name:     AthenaNamedQueryResource,
		Scope:    nuke.Account,
		Resource: &AthenaNamedQuery{},
		Lister:   &AthenaNamedQueryLister{},
	})
}

type AthenaNamedQueryLister struct{}

func (l *AthenaNamedQueryLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := athena.New(opts.Session)
	resources := make([]resource.Resource, 0)

	// List WorkGroup
	var workgroupNames []*string
	err := svc.ListWorkGroupsPages(
		&athena.ListWorkGroupsInput{},
		func(page *athena.ListWorkGroupsOutput, lastPage bool) bool {
			for _, workgroup := range page.WorkGroups {
				workgroupNames = append(workgroupNames, workgroup.Name)
			}
			return true
		},
	)
	if err != nil {
		return nil, err
	}

	// List NamedQueries or each WorkGroup
	var namedQueryIDs []*string
	for _, wgName := range workgroupNames {
		err := svc.ListNamedQueriesPages(
			&athena.ListNamedQueriesInput{WorkGroup: wgName},
			func(page *athena.ListNamedQueriesOutput, lastPage bool) bool {
				namedQueryIDs = append(namedQueryIDs, page.NamedQueryIds...)
				return true
			},
		)
		if err != nil {
			return nil, err
		}
	}

	// Create AthenaNamedQuery resource objects
	for _, id := range namedQueryIDs {
		resources = append(resources, &AthenaNamedQuery{
			svc: svc,
			id:  id,
		})
	}

	return resources, err
}

type AthenaNamedQuery struct {
	svc *athena.Athena
	id  *string
}

func (r *AthenaNamedQuery) Remove(_ context.Context) error {
	_, err := r.svc.DeleteNamedQuery(&athena.DeleteNamedQueryInput{
		NamedQueryId: r.id,
	})

	return err
}

func (r *AthenaNamedQuery) Properties() types.Properties {
	return types.NewProperties().
		Set("Id", *r.id)
}

func (r *AthenaNamedQuery) String() string {
	return *r.id
}
