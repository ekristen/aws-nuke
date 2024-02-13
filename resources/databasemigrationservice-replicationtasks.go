package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/databasemigrationservice"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const DatabaseMigrationServiceReplicationTaskResource = "DatabaseMigrationServiceReplicationTask"

func init() {
	registry.Register(&registry.Registration{
		Name:   DatabaseMigrationServiceReplicationTaskResource,
		Scope:  nuke.Account,
		Lister: &DatabaseMigrationServiceReplicationTaskLister{},
	})
}

type DatabaseMigrationServiceReplicationTaskLister struct{}

func (l *DatabaseMigrationServiceReplicationTaskLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := databasemigrationservice.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &databasemigrationservice.DescribeReplicationTasksInput{
		MaxRecords: aws.Int64(100),
	}

	for {
		output, err := svc.DescribeReplicationTasks(params)
		if err != nil {
			return nil, err
		}

		for _, replicationTask := range output.ReplicationTasks {
			resources = append(resources, &DatabaseMigrationServiceReplicationTask{
				svc: svc,
				ARN: replicationTask.ReplicationTaskArn,
			})
		}

		if output.Marker == nil {
			break
		}

		params.Marker = output.Marker
	}

	return resources, nil
}

type DatabaseMigrationServiceReplicationTask struct {
	svc *databasemigrationservice.DatabaseMigrationService
	ARN *string
}

func (f *DatabaseMigrationServiceReplicationTask) Remove(_ context.Context) error {
	_, err := f.svc.DeleteReplicationTask(&databasemigrationservice.DeleteReplicationTaskInput{
		ReplicationTaskArn: f.ARN,
	})

	return err
}

func (f *DatabaseMigrationServiceReplicationTask) String() string {
	return *f.ARN
}
