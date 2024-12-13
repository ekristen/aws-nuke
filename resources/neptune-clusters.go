package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/neptune"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const NeptuneClusterResource = "NeptuneCluster"

func init() {
	registry.Register(&registry.Registration{
		Name:     NeptuneClusterResource,
		Scope:    nuke.Account,
		Resource: &NeptuneCluster{},
		Lister:   &NeptuneClusterLister{},
	})
}

type NeptuneClusterLister struct{}

func (l *NeptuneClusterLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := neptune.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &neptune.DescribeDBClustersInput{
		MaxRecords: aws.Int64(100),
	}

	for {
		output, err := svc.DescribeDBClusters(params)
		if err != nil {
			return nil, err
		}

		for _, dbCluster := range output.DBClusters {
			var dbTags []*neptune.Tag
			tags, err := svc.ListTagsForResource(&neptune.ListTagsForResourceInput{
				ResourceName: dbCluster.DBClusterArn,
			})
			if err != nil {
				opts.Logger.WithError(err).Warn("failed to list tags for resource")
			} else {
				dbTags = tags.TagList
			}

			resources = append(resources, &NeptuneCluster{
				svc:  svc,
				ID:   dbCluster.DBClusterIdentifier,
				Tags: dbTags,
			})
		}

		if output.Marker == nil {
			break
		}

		params.Marker = output.Marker
	}

	return resources, nil
}

type NeptuneCluster struct {
	svc  *neptune.Neptune
	ID   *string
	Tags []*neptune.Tag
}

func (f *NeptuneCluster) Remove(_ context.Context) error {
	_, err := f.svc.DeleteDBCluster(&neptune.DeleteDBClusterInput{
		DBClusterIdentifier: f.ID,
		SkipFinalSnapshot:   aws.Bool(true),
	})

	return err
}

func (f *NeptuneCluster) Properties() types.Properties {
	return types.NewPropertiesFromStruct(f)
}

func (f *NeptuneCluster) String() string {
	return *f.ID
}
