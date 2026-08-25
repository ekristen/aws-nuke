package resources

import (
	"context"

	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/databasemigrationservice"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const DatabasemigrationserviceReplicationConfigsResource = "DatabaseMigrationServiceReplicationConfig"

func init() {
	registry.Register(&registry.Registration{
		Name:     DatabasemigrationserviceReplicationConfigsResource,
		Scope:    nuke.Account,
		Resource: &DatabasemigrationserviceReplicationConfigs{},
		Lister:   &DatabasemigrationserviceReplicationConfigsLister{},
	})
}

type DatabasemigrationserviceReplicationConfigsLister struct{}

func (l *DatabasemigrationserviceReplicationConfigsLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := databasemigrationservice.NewFromConfig(*opts.Config)
	var resources []resource.Resource

	params := &databasemigrationservice.DescribeReplicationConfigsInput{
		MaxRecords: aws.Int32(100),
	}
	for {
		output, err := svc.DescribeReplicationConfigs(ctx, params)
		if err != nil {
			return nil, err
		}

		for _, replicationConfig := range output.ReplicationConfigs {
			resources = append(resources, &DatabasemigrationserviceReplicationConfigs{
				svc:        svc,
				Arn:        replicationConfig.ReplicationConfigArn,
				Identifier: replicationConfig.ReplicationConfigIdentifier,
			})
		}

		if output.Marker == nil {
			break
		}

		params.Marker = output.Marker
	}
	return resources, nil
}

type DatabasemigrationserviceReplicationConfigs struct {
	svc        *databasemigrationservice.Client
	Arn        *string `description:"The ARN of the ReplicationConfig"`
	Identifier *string `description:"The Identifier of the ReplicationConfig"`
}

func (r *DatabasemigrationserviceReplicationConfigs) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteReplicationConfig(ctx, &databasemigrationservice.DeleteReplicationConfigInput{
		ReplicationConfigArn: r.Arn,
	})
	return err
}

func (r *DatabasemigrationserviceReplicationConfigs) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *DatabasemigrationserviceReplicationConfigs) String() string {
	return fmt.Sprintf("ReplicationConfigArn: %s", *r.Arn)
}
