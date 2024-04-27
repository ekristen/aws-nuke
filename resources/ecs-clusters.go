package resources

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/ecs/ecsiface"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/slices"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const ECSClusterResource = "ECSCluster"

func init() {
	registry.Register(&registry.Registration{
		Name:   ECSClusterResource,
		Scope:  nuke.Account,
		Lister: &ECSClusterLister{},
	})
}

type ECSClusterLister struct {
	mockSvc ecsiface.ECSAPI
}

func (l *ECSClusterLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	var svc ecsiface.ECSAPI
	if l.mockSvc != nil {
		svc = l.mockSvc
	} else {
		svc = ecs.New(opts.Session)
	}

	resources := make([]resource.Resource, 0)

	params := &ecs.ListClustersInput{
		MaxResults: aws.Int64(100),
	}

	for {
		output, err := svc.ListClusters(params)
		if err != nil {
			return nil, err
		}

		for _, clusterChunk := range slices.Chunk(output.ClusterArns, 100) {
			clusters, err := svc.DescribeClusters(&ecs.DescribeClustersInput{
				Clusters: clusterChunk,
			})
			if err != nil {
				logrus.WithError(err).Error("unable to retrieve clusters")
			}

			for _, cluster := range clusters.Clusters {
				resources = append(resources, &ECSCluster{
					svc:  svc,
					ARN:  cluster.ClusterArn,
					tags: cluster.Tags,
				})
			}
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type ECSCluster struct {
	svc  ecsiface.ECSAPI
	ARN  *string
	tags []*ecs.Tag
}

func (f *ECSCluster) Remove(_ context.Context) error {
	_, err := f.svc.DeleteCluster(&ecs.DeleteClusterInput{
		Cluster: f.ARN,
	})

	return err
}

func (f *ECSCluster) String() string {
	return *f.ARN
}

func (f *ECSCluster) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("ARN", f.ARN)

	for _, tag := range f.tags {
		properties.SetTag(tag.Key, tag.Value)
	}

	return properties
}
