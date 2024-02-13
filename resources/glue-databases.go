package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/glue"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const GlueDatabaseResource = "GlueDatabase"

func init() {
	registry.Register(&registry.Registration{
		Name:   GlueDatabaseResource,
		Scope:  nuke.Account,
		Lister: &GlueDatabaseLister{},
	})
}

type GlueDatabaseLister struct{}

func (l *GlueDatabaseLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := glue.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &glue.GetDatabasesInput{
		MaxResults: aws.Int64(100),
	}

	for {
		output, err := svc.GetDatabases(params)
		if err != nil {
			return nil, err
		}

		for _, database := range output.DatabaseList {
			resources = append(resources, &GlueDatabase{
				svc:  svc,
				name: database.Name,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type GlueDatabase struct {
	svc  *glue.Glue
	name *string
}

func (f *GlueDatabase) Remove(_ context.Context) error {
	_, err := f.svc.DeleteDatabase(&glue.DeleteDatabaseInput{
		Name: f.name,
	})

	return err
}

func (f *GlueDatabase) String() string {
	return *f.name
}
