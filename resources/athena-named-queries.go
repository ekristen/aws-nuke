package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/athena"

	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const AthenaNamedQueryResource = "AthenaNamedQuery"

func init() {
	resource.Register(&resource.Registration{
		Name:   AthenaNamedQueryResource,
		Scope:  nuke.Account,
		Lister: &AthenaNamedQueryLister{},
	})
}

type AthenaNamedQueryLister struct{}

type AthenaNamedQuery struct {
	svc *athena.Athena
	id  *string
}

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

func (a *AthenaNamedQuery) Remove(_ context.Context) error {
	_, err := a.svc.DeleteNamedQuery(&athena.DeleteNamedQueryInput{
		NamedQueryId: a.id,
	})

	return err
}

func (a *AthenaNamedQuery) Properties() types.Properties {
	return types.NewProperties().
		Set("Id", *a.id)
}

func (a *AthenaNamedQuery) String() string {
	return *a.id
}
