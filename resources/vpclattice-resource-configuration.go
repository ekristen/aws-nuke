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

const VPCLatticeResourceConfigurationResource = "VPCLatticeResourceConfiguration"

func init() {
	registry.Register(&registry.Registration{
		Name:     VPCLatticeResourceConfigurationResource,
		Scope:    nuke.Account,
		Resource: &VPCLatticeResourceConfiguration{},
		Lister:   &VPCLatticeResourceConfigurationLister{},
		DependsOn: []string{
			"VPCLatticeServiceNetworkResourceAssociation",
		},
	})
}

type VPCLatticeResourceConfigurationLister struct{}

func (l *VPCLatticeResourceConfigurationLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := vpclattice.NewFromConfig(*opts.Config)
	var resources []resource.Resource

	params := &vpclattice.ListResourceConfigurationsInput{}
	paginator := vpclattice.NewListResourceConfigurationsPaginator(svc, params)

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

			resources = append(resources, &VPCLatticeResourceConfiguration{
				svc:           svc,
				ID:            resp.Items[i].Id,
				Name:          resp.Items[i].Name,
				ARN:           resp.Items[i].Arn,
				Status:        string(resp.Items[i].Status),
				Type:          string(resp.Items[i].Type),
				AmazonManaged: resp.Items[i].AmazonManaged,
				CreatedAt:     resp.Items[i].CreatedAt,
				Tags:          tagResp.Tags,
			})
		}
	}

	return resources, nil
}

type VPCLatticeResourceConfiguration struct {
	svc           *vpclattice.Client
	ID            *string
	Name          *string
	ARN           *string
	Status        string
	Type          string
	AmazonManaged *bool
	CreatedAt     *time.Time
	Tags          map[string]string
}

func (r *VPCLatticeResourceConfiguration) Filter() error {
	if r.AmazonManaged != nil && *r.AmazonManaged {
		return fmt.Errorf("cannot delete Amazon managed resource")
	}
	switch vpclatticeTypes.ResourceConfigurationStatus(r.Status) {
	case vpclatticeTypes.ResourceConfigurationStatusDeleteInProgress,
		vpclatticeTypes.ResourceConfigurationStatusCreateInProgress,
		vpclatticeTypes.ResourceConfigurationStatusUpdateInProgress:
		return fmt.Errorf("resource configuration is in %s state", r.Status)
	}
	return nil
}

func (r *VPCLatticeResourceConfiguration) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteResourceConfiguration(ctx, &vpclattice.DeleteResourceConfigurationInput{
		ResourceConfigurationIdentifier: r.ID,
	})
	return err
}

func (r *VPCLatticeResourceConfiguration) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *VPCLatticeResourceConfiguration) String() string {
	return *r.Name
}
