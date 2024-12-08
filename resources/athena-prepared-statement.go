package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/aws/aws-sdk-go/service/athena"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const AthenaPreparedStatementResource = "AthenaPreparedStatement"

func init() {
	registry.Register(&registry.Registration{
		Name:     AthenaPreparedStatementResource,
		Scope:    nuke.Account,
		Resource: &AthenaPreparedStatement{},
		Lister:   &AthenaPreparedStatementLister{},
	})
}

type AthenaPreparedStatementLister struct{}

func (l *AthenaPreparedStatementLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := athena.New(opts.Session)
	resources := make([]resource.Resource, 0)

	workgroups, err := svc.ListWorkGroups(&athena.ListWorkGroupsInput{})
	if err != nil {
		return nil, err
	}

	for _, workgroup := range workgroups.WorkGroups {
		params := &athena.ListPreparedStatementsInput{
			WorkGroup:  workgroup.Name,
			MaxResults: aws.Int64(50),
		}

		for {
			output, err := svc.ListPreparedStatements(params)
			if err != nil {
				return nil, err
			}

			for _, statement := range output.PreparedStatements {
				resources = append(resources, &AthenaPreparedStatement{
					svc:       svc,
					Name:      statement.StatementName,
					WorkGroup: workgroup.Name,
				})
			}

			if output.NextToken == nil {
				break
			}

			params.NextToken = output.NextToken
		}
	}

	return resources, nil
}

type AthenaPreparedStatement struct {
	svc       *athena.Athena
	Name      *string
	WorkGroup *string
}

func (r *AthenaPreparedStatement) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *AthenaPreparedStatement) Remove(_ context.Context) error {
	_, err := r.svc.DeletePreparedStatement(&athena.DeletePreparedStatementInput{
		StatementName: r.Name,
		WorkGroup:     r.WorkGroup,
	})

	return err
}

func (r *AthenaPreparedStatement) String() string {
	return *r.Name
}
