package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/databasemigrationservice"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const DatabaseMigrationServiceReplicationInstanceResource = "DatabaseMigrationServiceReplicationInstance"

func init() {
	registry.Register(&registry.Registration{
		Name:   DatabaseMigrationServiceReplicationInstanceResource,
		Scope:  nuke.Account,
		Lister: &DatabaseMigrationServiceReplicationInstanceLister{},
	})
}

type DatabaseMigrationServiceReplicationInstanceLister struct{}

func (l *DatabaseMigrationServiceReplicationInstanceLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := databasemigrationservice.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &databasemigrationservice.DescribeReplicationInstancesInput{
		MaxRecords: aws.Int64(100),
	}

	for {
		output, err := svc.DescribeReplicationInstances(params)
		if err != nil {
			return nil, err
		}

		for _, replicationInstance := range output.ReplicationInstances {
			resources = append(resources, &DatabaseMigrationServiceReplicationInstance{
				svc: svc,
				ARN: replicationInstance.ReplicationInstanceArn,
			})
		}

		if output.Marker == nil {
			break
		}

		params.Marker = output.Marker
	}

	return resources, nil
}

type DatabaseMigrationServiceReplicationInstance struct {
	svc *databasemigrationservice.DatabaseMigrationService
	ARN *string
}

func (f *DatabaseMigrationServiceReplicationInstance) Remove(_ context.Context) error {
	_, err := f.svc.DeleteReplicationInstance(&databasemigrationservice.DeleteReplicationInstanceInput{
		ReplicationInstanceArn: f.ARN,
	})

	return err
}

func (f *DatabaseMigrationServiceReplicationInstance) String() string {
	return *f.ARN
}
