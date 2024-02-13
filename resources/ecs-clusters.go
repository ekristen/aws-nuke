package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

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

type ECSClusterLister struct{}

func (l *ECSClusterLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := ecs.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &ecs.ListClustersInput{
		MaxResults: aws.Int64(100),
	}

	for {
		output, err := svc.ListClusters(params)
		if err != nil {
			return nil, err
		}

		for _, clusterArn := range output.ClusterArns {
			resources = append(resources, &ECSCluster{
				svc: svc,
				ARN: clusterArn,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type ECSCluster struct {
	svc *ecs.ECS
	ARN *string
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
