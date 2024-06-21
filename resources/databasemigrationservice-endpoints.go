package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/databasemigrationservice"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const DatabaseMigrationServiceEndpointResource = "DatabaseMigrationServiceEndpoint"

func init() {
	registry.Register(&registry.Registration{
		Name:   DatabaseMigrationServiceEndpointResource,
		Scope:  nuke.Account,
		Lister: &DatabaseMigrationServiceEndpointLister{},
	})
}

type DatabaseMigrationServiceEndpointLister struct{}

func (l *DatabaseMigrationServiceEndpointLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := databasemigrationservice.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &databasemigrationservice.DescribeEndpointsInput{
		MaxRecords: aws.Int64(100),
	}

	for {
		output, err := svc.DescribeEndpoints(params)
		if err != nil {
			return nil, err
		}

		for _, endpoint := range output.Endpoints {
			resources = append(resources, &DatabaseMigrationServiceEndpoint{
				svc: svc,
				ARN: endpoint.EndpointArn,
			})
		}

		if output.Marker == nil {
			break
		}

		params.Marker = output.Marker
	}

	return resources, nil
}

type DatabaseMigrationServiceEndpoint struct {
	svc *databasemigrationservice.DatabaseMigrationService
	ARN *string
}

func (f *DatabaseMigrationServiceEndpoint) Remove(_ context.Context) error {
	_, err := f.svc.DeleteEndpoint(&databasemigrationservice.DeleteEndpointInput{
		EndpointArn: f.ARN,
	})

	return err
}

func (f *DatabaseMigrationServiceEndpoint) String() string {
	return *f.ARN
}
