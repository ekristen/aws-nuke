package resources

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws"         //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/ecs" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

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

		for {
			output, err := svc.ListServices(serviceParams)
			if err != nil {
				return nil, err
			}

			for _, serviceArn := range output.ServiceArns {
				ecsService := &ECSService{
					svc:        svc,
					ServiceARN: serviceArn,
					ClusterARN: clusterArn,
				}

				// Fetch tags for the service
				tags, err := svc.ListTagsForResource(&ecs.ListTagsForResourceInput{
					ResourceArn: serviceArn,
				})
				if err != nil {
					logrus.WithError(err).Error("unable to get tags for ECS service")
				} else if tags != nil {
					ecsService.Tags = tags.Tags
				}

				resources = append(resources, ecsService)
			}

			if output.NextToken == nil {
				break
			}

			serviceParams.NextToken = output.NextToken
		}
	}

	return resources, nil
}

type ECSService struct {
	svc        *ecs.ECS
	ServiceARN *string    `description:"The ARN of the ECS service"`
	ClusterARN *string    `description:"The ARN of the ECS cluster"`
	Tags       []*ecs.Tag `description:"The tags associated with the service"`
}

func (f *ECSService) Properties() types.Properties {
	return types.NewPropertiesFromStruct(f)
}

func (f *ECSService) Remove(_ context.Context) error {
	_, err := f.svc.DeleteService(&ecs.DeleteServiceInput{
		Cluster: f.ClusterARN,
		Service: f.ServiceARN,
		Force:   aws.Bool(true),
	})

	return err
}

func (f *ECSService) String() string {
	return fmt.Sprintf("%s -> %s", *f.ServiceARN, *f.ClusterARN)
}
