package resources

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/vpclattice"
	vpclatticeTypes "github.com/aws/aws-sdk-go-v2/service/vpclattice/types"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const VPCLatticeServiceNetworkResourceAssociationResource = "VPCLatticeServiceNetworkResourceAssociation"

func init() {
	registry.Register(&registry.Registration{
		Name:     VPCLatticeServiceNetworkResourceAssociationResource,
		Scope:    nuke.Account,
		Resource: &VPCLatticeServiceNetworkResourceAssociation{},
		Lister:   &VPCLatticeServiceNetworkResourceAssociationLister{},
	})
}

type VPCLatticeServiceNetworkResourceAssociationLister struct{}

func (l *VPCLatticeServiceNetworkResourceAssociationLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := vpclattice.NewFromConfig(*opts.Config)
	var resources []resource.Resource

	params := &vpclattice.ListServiceNetworkResourceAssociationsInput{}
	paginator := vpclattice.NewListServiceNetworkResourceAssociationsPaginator(svc, params)

	for paginator.HasMorePages() {
		resp, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for i := range resp.Items {
			tagResp, err := svc.ListTagsForResource(ctx, &vpclattice.ListTagsForResourceInput{
				ResourceArn: resp.Items[i].Arn,
			})
			if err != nil {
				return nil, err
			}

			resources = append(resources, &VPCLatticeServiceNetworkResourceAssociation{
				svc:                       svc,
				ID:                        resp.Items[i].Id,
				ARN:                       resp.Items[i].Arn,
				Status:                    string(resp.Items[i].Status),
				ServiceNetworkName:        resp.Items[i].ServiceNetworkName,
				ResourceConfigurationName: resp.Items[i].ResourceConfigurationName,
				IsManagedAssociation:      resp.Items[i].IsManagedAssociation,
				CreatedAt:                 resp.Items[i].CreatedAt,
				Tags:                      tagResp.Tags,
			})
		}
	}

	return resources, nil
}

type VPCLatticeServiceNetworkResourceAssociation struct {
	svc                       *vpclattice.Client
	ID                        *string
	ARN                       *string
	Status                    string
	ServiceNetworkName        *string
	ResourceConfigurationName *string
	IsManagedAssociation      *bool
	CreatedAt                 *time.Time
	Tags                      map[string]string
}

func (r *VPCLatticeServiceNetworkResourceAssociation) Filter() error {
	if r.IsManagedAssociation != nil && *r.IsManagedAssociation {
		return fmt.Errorf("cannot delete managed association")
	}
	switch vpclatticeTypes.ServiceNetworkResourceAssociationStatus(r.Status) {
	case vpclatticeTypes.ServiceNetworkResourceAssociationStatusDeleteInProgress,
		vpclatticeTypes.ServiceNetworkResourceAssociationStatusCreateInProgress:
		return fmt.Errorf("service network resource association is in %s state", r.Status)
	}
	return nil
}

func (r *VPCLatticeServiceNetworkResourceAssociation) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteServiceNetworkResourceAssociation(ctx, &vpclattice.DeleteServiceNetworkResourceAssociationInput{
		ServiceNetworkResourceAssociationIdentifier: r.ID,
	})
	return err
}

func (r *VPCLatticeServiceNetworkResourceAssociation) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *VPCLatticeServiceNetworkResourceAssociation) String() string {
	return *r.ID
}
