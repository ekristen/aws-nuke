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

const ECSServiceResource = "ECSService"

func init() {
	registry.Register(&registry.Registration{
		Name:     ECSServiceResource,
		Scope:    nuke.Account,
		Resource: &ECSService{},
		Lister:   &ECSServiceLister{},
	})
}

type ECSServiceLister struct{}

func (l *ECSServiceLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := ecs.New(opts.Session)
	resources := make([]resource.Resource, 0)
	clusters := []*string{}

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
		serviceParams := &ecs.ListServicesInput{
			Cluster:    clusterArn,
			MaxResults: aws.Int64(10),
		}
		output, err := svc.ListServices(serviceParams)
		if err != nil {
			return nil, err
		}

		for _, serviceArn := range output.ServiceArns {
			resources = append(resources, &ECSService{
				svc:        svc,
				serviceARN: serviceArn,
				clusterARN: clusterArn,
			})
		}

		if output.NextToken == nil {
			continue
		}

		serviceParams.NextToken = output.NextToken
	}

	return resources, nil
}

type ECSService struct {
	svc        *ecs.ECS
	serviceARN *string
	clusterARN *string
}

func (f *ECSService) Remove(_ context.Context) error {
	_, err := f.svc.DeleteService(&ecs.DeleteServiceInput{
		Cluster: f.clusterARN,
		Service: f.serviceARN,
		Force:   aws.Bool(true),
	})

	return err
}

func (f *ECSService) String() string {
	return fmt.Sprintf("%s -> %s", *f.serviceARN, *f.clusterARN)
}
