package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const RDSDBClusterResource = "RDSDBCluster"

func init() {
	registry.Register(&registry.Registration{
		Name:   RDSDBClusterResource,
		Scope:  nuke.Account,
		Lister: &RDSDBClusterLister{},
		DeprecatedAliases: []string{
			"RDSCluster",
		},
	})
}

type RDSDBClusterLister struct{}

func (l *RDSDBClusterLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := rds.New(opts.Session)

	params := &rds.DescribeDBClustersInput{}
	resp, err := svc.DescribeDBClusters(params)
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, 0)
	for _, instance := range resp.DBClusters {
		tags, err := svc.ListTagsForResource(&rds.ListTagsForResourceInput{
			ResourceName: instance.DBClusterArn,
		})

		if err != nil {
			continue
		}

		resources = append(resources, &RDSDBCluster{
			svc:                svc,
			id:                 *instance.DBClusterIdentifier,
			deletionProtection: *instance.DeletionProtection,
			tags:               tags.TagList,
		})
	}

	return resources, nil
}

type RDSDBCluster struct {
	svc                *rds.RDS
	id                 string
	deletionProtection bool
	tags               []*rds.Tag
}

func (i *RDSDBCluster) Remove(_ context.Context) error {
	if i.deletionProtection {
		modifyParams := &rds.ModifyDBClusterInput{
			DBClusterIdentifier: &i.id,
			DeletionProtection:  aws.Bool(false),
		}
		_, err := i.svc.ModifyDBCluster(modifyParams)
		if err != nil {
			return err
		}
	}

	params := &rds.DeleteDBClusterInput{
		DBClusterIdentifier: &i.id,
		SkipFinalSnapshot:   aws.Bool(true),
	}

	_, err := i.svc.DeleteDBCluster(params)
	if err != nil {
		return err
	}

	return nil
}

func (i *RDSDBCluster) String() string {
	return i.id
}

func (i *RDSDBCluster) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("Identifier", i.id)
	properties.Set("Deletion Protection", i.deletionProtection)

	for _, tag := range i.tags {
		properties.SetTag(tag.Key, tag.Value)
	}

	return properties
}
