package resources

import (
	"context"

	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const ECSClusterInstanceResource = "ECSClusterInstance"

func init() {
	registry.Register(&registry.Registration{
		Name:   ECSClusterInstanceResource,
		Scope:  nuke.Account,
		Lister: &ECSClusterInstanceLister{},
	})
}

type ECSClusterInstanceLister struct{}

func (l *ECSClusterInstanceLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := ecs.New(opts.Session)
	resources := make([]resource.Resource, 0)
	var clusters []*string

	clusterParams := &ecs.ListClustersInput{
		MaxResults: aws.Int64(100),
	}

	// Iterate over clusters to ensure we dont presume its always default associations
	for {
		output, err := svc.ListClusters(clusterParams)
		if err != nil {
			return nil, err
		}

		clusters = append(clusters, output.ClusterArns...)

		if output.NextToken == nil {
			break
		}

		clusterParams.NextToken = output.NextToken
	}

	// Iterate over known clusters and discover their instances
	// to prevent assuming default is always used
	for _, clusterArn := range clusters {
		instanceParams := &ecs.ListContainerInstancesInput{
			Cluster:    clusterArn,
			MaxResults: aws.Int64(100),
		}
		output, err := svc.ListContainerInstances(instanceParams)
		if err != nil {
			return nil, err
		}

		for _, instanceArn := range output.ContainerInstanceArns {
			resources = append(resources, &ECSClusterInstance{
				svc:         svc,
				instanceARN: instanceArn,
				clusterARN:  clusterArn,
			})
		}

		if output.NextToken == nil {
			break
		}

		instanceParams.NextToken = output.NextToken
	}

	return resources, nil
}

type ECSClusterInstance struct {
	svc         *ecs.ECS
	instanceARN *string
	clusterARN  *string
}

func (f *ECSClusterInstance) Remove(_ context.Context) error {
	_, err := f.svc.DeregisterContainerInstance(&ecs.DeregisterContainerInstanceInput{
		Cluster:           f.clusterARN,
		ContainerInstance: f.instanceARN,
		Force:             aws.Bool(true),
	})

	return err
}

func (f *ECSClusterInstance) String() string {
	return fmt.Sprintf("%s -> %s", *f.instanceARN, *f.clusterARN)
}
