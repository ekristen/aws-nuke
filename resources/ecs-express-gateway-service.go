package resources

import (
	"context"

	"github.com/gotidy/ptr"
	"github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go-v2/service/ecs"
	ecstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

// ECSExpressGatewayServiceResource is the resource type identifier for ECS services
// created via Express Mode (ResourceManagementType=ECS). These services are distinct
// from standard ECS services and must be deleted using the DeleteExpressGatewayService
// API — calling DeleteService on them results in an InvalidParameterException.
const ECSExpressGatewayServiceResource = "ECSExpressGatewayService"

func init() {
	registry.Register(&registry.Registration{
		Name:     ECSExpressGatewayServiceResource,
		Scope:    nuke.Account,
		Resource: &ECSExpressGatewayService{},
		Lister:   &ECSExpressGatewayServiceLister{},
	})
}

// ECSExpressGatewayServiceLister lists only ECS services created via Express Mode.
// Express Mode services have ResourceManagementType=ECS and are created through
// CreateExpressGatewayService rather than the standard CreateService API. They
// appear in ListServices results alongside standard services, so the
// ResourceManagementType filter is used to scope results to Express Mode only.
type ECSExpressGatewayServiceLister struct {
	mockSvc ECSServiceClient
}

func (l *ECSExpressGatewayServiceLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	var svc ECSServiceClient
	if l.mockSvc != nil {
		svc = l.mockSvc
	} else {
		svc = ecs.NewFromConfig(*opts.Config)
	}

	clusters, err := l.listClusters(ctx, svc)
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, 0)

	for _, clusterArn := range clusters {
		serviceArns, err := l.listExpressServices(ctx, svc, clusterArn)
		if err != nil {
			return nil, err
		}

		for _, serviceArn := range serviceArns {
			ecsExpressSvc := &ECSExpressGatewayService{
				svc:        svc,
				ServiceARN: ptr.String(serviceArn),
				Tags:       make(map[string]string),
			}

			tags, err := svc.ListTagsForResource(ctx, &ecs.ListTagsForResourceInput{
				ResourceArn: ptr.String(serviceArn),
			})
			if err != nil {
				logrus.WithError(err).Error("unable to get tags for ECS express gateway service")
			} else if tags != nil {
				for _, tag := range tags.Tags {
					if tag.Key != nil && tag.Value != nil {
						ecsExpressSvc.Tags[*tag.Key] = *tag.Value
					}
				}
			}

			resources = append(resources, ecsExpressSvc)
		}
	}

	return resources, nil
}

// listClusters returns all ECS cluster ARNs in the account, paginating as needed.
func (l *ECSExpressGatewayServiceLister) listClusters(ctx context.Context, svc ECSServiceClient) ([]string, error) {
	var clusters []string

	params := &ecs.ListClustersInput{
		MaxResults: ptr.Int32(100),
	}

	for {
		output, err := svc.ListClusters(ctx, params)
		if err != nil {
			return nil, err
		}

		clusters = append(clusters, output.ClusterArns...)

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return clusters, nil
}

// listExpressServices returns all Express Mode service ARNs within the given cluster,
// paginating as needed. Only services with ResourceManagementType=ECS are returned;
// standard CUSTOMER-managed services are excluded at the API level via the
// ResourceManagementType filter, avoiding any need for a DescribeServices call.
func (l *ECSExpressGatewayServiceLister) listExpressServices(ctx context.Context, svc ECSServiceClient, clstrArn string) ([]string, error) {
	var serviceArns []string

	params := &ecs.ListServicesInput{
		Cluster:                ptr.String(clstrArn),
		MaxResults:             ptr.Int32(10),
		ResourceManagementType: ecstypes.ResourceManagementTypeEcs,
	}

	for {
		output, err := svc.ListServices(ctx, params)
		if err != nil {
			return nil, err
		}

		serviceArns = append(serviceArns, output.ServiceArns...)

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return serviceArns, nil
}

// ECSExpressGatewayService represents an ECS service created via Express Mode.
// Express Mode services are fully AWS-managed: ECS automatically provisions and
// manages the ALB, target groups, security groups, and auto-scaling policies.
// Deletion requires DeleteExpressGatewayService (not DeleteService) and does not
// accept a Cluster or Force parameter — only the ServiceArn is required.
type ECSExpressGatewayService struct {
	svc        ECSServiceClient
	ServiceARN *string           `description:"The ARN of the ECS Express Gateway service"`
	Tags       map[string]string `description:"The tags associated with the service"`
}

func (r *ECSExpressGatewayService) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

// Remove deletes the Express Mode ECS service and all AWS-managed infrastructure
// associated with it (ALB, target groups, security groups, auto-scaling policies).
// DeleteExpressGatewayService is used here because DeleteService is not valid for
// services with ResourceManagementType=ECS and will return an InvalidParameterException.
func (r *ECSExpressGatewayService) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteExpressGatewayService(ctx, &ecs.DeleteExpressGatewayServiceInput{
		ServiceArn: r.ServiceARN,
	})
	return err
}

func (r *ECSExpressGatewayService) String() string {
	return *r.ServiceARN
}
