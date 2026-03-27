package resources

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/connect"
	connecttypes "github.com/aws/aws-sdk-go-v2/service/connect/types"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const ConnectInstanceResource = "ConnectInstance"

func init() {
	registry.Register(&registry.Registration{
		Name:     ConnectInstanceResource,
		Scope:    nuke.Account,
		Resource: &ConnectInstance{},
		Lister:   &ConnectInstanceLister{},
		DependsOn: []string{
			ConnectContactFlowResource,
			ConnectContactFlowModuleResource,
			ConnectQueueResource,
			ConnectRoutingProfileResource,
			ConnectUserResource,
			ConnectSecurityProfileResource,
			ConnectPhoneNumberResource,
			ConnectHoursOfOperationResource,
			ConnectQuickConnectResource,
			ConnectRuleResource,
			ConnectIntegrationAssociationResource,
		},
	})
}

type ConnectInstanceLister struct{}

func (l *ConnectInstanceLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := connect.NewFromConfig(*opts.Config)
	var resources []resource.Resource

	instances, err := listConnectInstances(ctx, svc)
	if err != nil {
		return nil, err
	}

	for _, instance := range instances {
		var tags map[string]string
		if instance.Arn != nil {
			tagsResp, err := svc.ListTagsForResource(ctx, &connect.ListTagsForResourceInput{
				ResourceArn: instance.Arn,
			})
			if err != nil {
				opts.Logger.Warnf("unable to fetch tags for connect instance: %s", *instance.Arn)
			} else {
				tags = tagsResp.Tags
			}
		}

		resources = append(resources, &ConnectInstance{
			svc:           svc,
			ID:            instance.Id,
			InstanceAlias: instance.InstanceAlias,
			ARN:           instance.Arn,
			Status:        string(instance.InstanceStatus),
			CreatedAt:     instance.CreatedTime,
			Tags:          tags,
		})
	}

	return resources, nil
}

type ConnectInstance struct {
	svc           *connect.Client
	ID            *string
	InstanceAlias *string
	ARN           *string
	Status        string
	CreatedAt     *time.Time
	Tags          map[string]string
}

func (r *ConnectInstance) Filter() error {
	if r.Status == string(connecttypes.InstanceStatusCreationInProgress) {
		return fmt.Errorf("instance is being created")
	}
	if r.Status == string(connecttypes.InstanceStatusCreationFailed) {
		return fmt.Errorf("instance creation failed")
	}
	return nil
}

func (r *ConnectInstance) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteInstance(ctx, &connect.DeleteInstanceInput{
		InstanceId: r.ID,
	})
	return err
}

func (r *ConnectInstance) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *ConnectInstance) String() string {
	if r.InstanceAlias != nil {
		return *r.InstanceAlias
	}
	return *r.ID
}
