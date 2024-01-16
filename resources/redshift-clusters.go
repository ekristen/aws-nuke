package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/redshift"

	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const RedshiftClusterResource = "RedshiftCluster"

func init() {
	resource.Register(resource.Registration{
		Name:   RedshiftClusterResource,
		Scope:  nuke.Account,
		Lister: &RedshiftClusterLister{},
	})
}

type RedshiftClusterLister struct{}

func (l *RedshiftClusterLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := redshift.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &redshift.DescribeClustersInput{
		MaxRecords: aws.Int64(100),
	}

	for {
		output, err := svc.DescribeClusters(params)
		if err != nil {
			return nil, err
		}

		for _, cluster := range output.Clusters {
			resources = append(resources, &RedshiftCluster{
				svc:     svc,
				cluster: cluster,
			})
		}

		if output.Marker == nil {
			break
		}

		params.Marker = output.Marker
	}

	return resources, nil
}

type RedshiftCluster struct {
	svc     *redshift.Redshift
	cluster *redshift.Cluster
}

func (f *RedshiftCluster) Properties() types.Properties {
	properties := types.NewProperties().
		Set("CreatedTime", f.cluster.ClusterCreateTime)

	for _, tag := range f.cluster.Tags {
		properties.SetTag(tag.Key, tag.Value)
	}

	return properties
}

func (f *RedshiftCluster) Remove(_ context.Context) error {

	_, err := f.svc.DeleteCluster(&redshift.DeleteClusterInput{
		ClusterIdentifier:        f.cluster.ClusterIdentifier,
		SkipFinalClusterSnapshot: aws.Bool(true),
	})

	return err
}

func (f *RedshiftCluster) String() string {
	return *f.cluster.ClusterIdentifier
}
