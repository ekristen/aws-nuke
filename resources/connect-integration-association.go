package resources

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/connect"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const ConnectIntegrationAssociationResource = "ConnectIntegrationAssociation"

func init() {
	registry.Register(&registry.Registration{
		Name:     ConnectIntegrationAssociationResource,
		Scope:    nuke.Account,
		Resource: &ConnectIntegrationAssociation{},
		Lister:   &ConnectIntegrationAssociationLister{},
	})
}

type ConnectIntegrationAssociationLister struct{}

func (l *ConnectIntegrationAssociationLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := connect.NewFromConfig(*opts.Config)
	var resources []resource.Resource

	instances, err := listConnectInstances(ctx, svc)
	if err != nil {
		return nil, err
	}

	for _, instance := range instances {
		paginator := connect.NewListIntegrationAssociationsPaginator(svc, &connect.ListIntegrationAssociationsInput{
			InstanceId: instance.Id,
		})

		for paginator.HasMorePages() {
			resp, err := paginator.NextPage(ctx)
			if err != nil {
				return nil, err
			}

			for _, assoc := range resp.IntegrationAssociationSummaryList {
				resources = append(resources, &ConnectIntegrationAssociation{
					svc:                      svc,
					InstanceID:               assoc.InstanceId,
					IntegrationAssociationID: assoc.IntegrationAssociationId,
					IntegrationType:          string(assoc.IntegrationType),
					IntegrationARN:           assoc.IntegrationArn,
					SourceApplicationName:    assoc.SourceApplicationName,
				})
			}
		}
	}

	return resources, nil
}

type ConnectIntegrationAssociation struct {
	svc                      *connect.Client
	InstanceID               *string
	IntegrationAssociationID *string
	IntegrationType          string
	IntegrationARN           *string
	SourceApplicationName    *string
}

func (r *ConnectIntegrationAssociation) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteIntegrationAssociation(ctx, &connect.DeleteIntegrationAssociationInput{
		InstanceId:               r.InstanceID,
		IntegrationAssociationId: r.IntegrationAssociationID,
	})
	return err
}

func (r *ConnectIntegrationAssociation) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *ConnectIntegrationAssociation) String() string {
	if r.SourceApplicationName != nil {
		return *r.SourceApplicationName
	}
	return *r.IntegrationAssociationID
}
