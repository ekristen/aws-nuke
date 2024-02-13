package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elasticache"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const ElasticacheReplicationGroupResource = "ElasticacheReplicationGroup"

func init() {
	registry.Register(&registry.Registration{
		Name:   ElasticacheReplicationGroupResource,
		Scope:  nuke.Account,
		Lister: &ElasticacheReplicationGroupLister{},
	})
}

type ElasticacheReplicationGroupLister struct{}

func (l *ElasticacheReplicationGroupLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := elasticache.New(opts.Session)
	var resources []resource.Resource

	params := &elasticache.DescribeReplicationGroupsInput{MaxRecords: aws.Int64(100)}

	for {
		resp, err := svc.DescribeReplicationGroups(params)
		if err != nil {
			return nil, err
		}

		for _, replicationGroup := range resp.ReplicationGroups {
			resources = append(resources, &ElasticacheReplicationGroup{
				svc:     svc,
				groupID: replicationGroup.ReplicationGroupId,
			})
		}

		if resp.Marker == nil {
			break
		}

		params.Marker = resp.Marker
	}

	return resources, nil
}

type ElasticacheReplicationGroup struct {
	svc     *elasticache.ElastiCache
	groupID *string
}

func (i *ElasticacheReplicationGroup) Remove(_ context.Context) error {
	params := &elasticache.DeleteReplicationGroupInput{
		ReplicationGroupId: i.groupID,
	}

	_, err := i.svc.DeleteReplicationGroup(params)
	if err != nil {
		return err
	}

	return nil
}

func (i *ElasticacheReplicationGroup) String() string {
	return *i.groupID
}
