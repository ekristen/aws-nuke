package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/databasemigrationservice"

	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const DatabaseMigrationServiceSubnetGroupResource = "DatabaseMigrationServiceSubnetGroup"

func init() {
	resource.Register(&resource.Registration{
		Name:   DatabaseMigrationServiceSubnetGroupResource,
		Scope:  nuke.Account,
		Lister: &DatabaseMigrationServiceSubnetGroupLister{},
	})
}

type DatabaseMigrationServiceSubnetGroupLister struct{}

func (l *DatabaseMigrationServiceSubnetGroupLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := databasemigrationservice.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &databasemigrationservice.DescribeReplicationSubnetGroupsInput{
		MaxRecords: aws.Int64(100),
	}

	for {
		output, err := svc.DescribeReplicationSubnetGroups(params)
		if err != nil {
			return nil, err
		}

		for _, replicationSubnetGroup := range output.ReplicationSubnetGroups {
			resources = append(resources, &DatabaseMigrationServiceSubnetGroup{
				svc: svc,
				ID:  replicationSubnetGroup.ReplicationSubnetGroupIdentifier,
			})
		}

		if output.Marker == nil {
			break
		}

		params.Marker = output.Marker
	}

	return resources, nil
}

type DatabaseMigrationServiceSubnetGroup struct {
	svc *databasemigrationservice.DatabaseMigrationService
	ID  *string
}

func (f *DatabaseMigrationServiceSubnetGroup) Remove(_ context.Context) error {
	_, err := f.svc.DeleteReplicationSubnetGroup(&databasemigrationservice.DeleteReplicationSubnetGroupInput{
		ReplicationSubnetGroupIdentifier: f.ID,
	})

	return err
}

func (f *DatabaseMigrationServiceSubnetGroup) String() string {
	return *f.ID
}
