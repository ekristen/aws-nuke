package resources

import (
	"context"
	"fmt"

	"github.com/gotidy/ptr"
	"github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go-v2/service/ecs"
	ecstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"

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

type ECSServiceLister struct {
	mockSvc ECSServiceClient
}

func (l *ECSServiceLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	var svc ECSServiceClient
	if l.mockSvc != nil {
		svc = l.mockSvc
	} else {
		svc = ecs.NewFromConfig(*opts.Config)
	}

	resources := make([]resource.Resource, 0)
	clusters := []string{}

	clusterParams := &ecs.ListClustersInput{
		MaxResults: ptr.Int32(100),
	}

	// Iterate over clusters to ensure we dont presume its always default associations
	for {
		output, err := svc.ListClusters(ctx, clusterParams)
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
			Cluster:                ptr.String(clusterArn),
			MaxResults:             ptr.Int32(10),
			ResourceManagementType: ecstypes.ResourceManagementTypeCustomer,
		}

		for {
			output, err := svc.ListServices(ctx, serviceParams)
			if err != nil {
				return nil, err
			}

			for _, serviceArn := range output.ServiceArns {
				ecsService := &ECSService{
					svc:        svc,
					ServiceARN: ptr.String(serviceArn),
					ClusterARN: ptr.String(clusterArn),
					Tags:       make(map[string]string),
				}

				// Fetch tags for the service
				tags, err := svc.ListTagsForResource(ctx, &ecs.ListTagsForResourceInput{
					ResourceArn: ptr.String(serviceArn),
				})
				if err != nil {
					logrus.WithError(err).Error("unable to get tags for ECS service")
				} else if tags != nil {
					for _, tag := range tags.Tags {
						if tag.Key != nil && tag.Value != nil {
							ecsService.Tags[*tag.Key] = *tag.Value
						}
					}
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
	svc        ECSServiceClient
	ServiceARN *string           `description:"The ARN of the ECS service"`
	ClusterARN *string           `description:"The ARN of the ECS cluster"`
	Tags       map[string]string `description:"The tags associated with the service"`
}

func (f *ECSService) Properties() types.Properties {
	return types.NewPropertiesFromStruct(f)
}

func (f *ECSService) Remove(ctx context.Context) error {
	_, err := f.svc.DeleteService(ctx, &ecs.DeleteServiceInput{
		Cluster: f.ClusterARN,
		Service: f.ServiceARN,
		Force:   ptr.Bool(true),
	})

	return err
}

func (f *ECSService) String() string {
	return fmt.Sprintf("%s -> %s", *f.ServiceARN, *f.ClusterARN)
}
