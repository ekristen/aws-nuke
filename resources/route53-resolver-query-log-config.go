package resources

import (
	"context"
	"errors"

	r53r "github.com/aws/aws-sdk-go-v2/service/route53resolver"
	r53rtypes "github.com/aws/aws-sdk-go-v2/service/route53resolver/types"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const Route53ResolverQueryLogConfigResource = "Route53ResolverQueryLogConfig"

func init() {
	registry.Register(&registry.Registration{
		Name:     Route53ResolverQueryLogConfigResource,
		Scope:    nuke.Account,
		Resource: &Route53ResolverQueryLogConfig{},
		Lister:   &Route53ResolverQueryLogConfigLister{},
	})
}

type Route53ResolverQueryLogConfigLister struct {
	svc Route53ResolverAPI
}

// List returns a list of all Route53 Resolver query log configs before filtering to be nuked
func (l *Route53ResolverQueryLogConfigLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	var resources []resource.Resource

	if l.svc == nil {
		l.svc = r53r.NewFromConfig(*opts.Config)
	}

	resourceAssociations, vpcErr := qlcsToAssociationIds(ctx, l.svc)
	if vpcErr != nil {
		return nil, vpcErr
	}

	params := &r53r.ListResolverQueryLogConfigsInput{}
	for {
		resp, err := l.svc.ListResolverQueryLogConfigs(ctx, params)
		if err != nil {
			return nil, err
		}

		for _, qlc := range resp.ResolverQueryLogConfigs {
			resources = append(resources, &Route53ResolverQueryLogConfig{
				svc:                    l.svc,
				resourceAssociationIds: resourceAssociations[*qlc.Id],
				Arn:                    qlc.Arn,
				AssociationCount:       qlc.AssociationCount,
				CreationTime:           qlc.CreationTime,
				CreatorRequestID:       qlc.CreatorRequestId,
				DestinationArn:         qlc.DestinationArn,
				ID:                     qlc.Id,
				OwnerID:                qlc.OwnerId,
				Name:                   qlc.Name,
				ShareStatus:            qlc.ShareStatus,
				Status:                 qlc.Status,
			})
		}

		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return resources, nil
}

// Route53ResolverQueryLogConfig is the resource type
type Route53ResolverQueryLogConfig struct {
	svc                    Route53ResolverAPI
	resourceAssociationIds []*string
	Arn                    *string
	AssociationCount       int32
	CreationTime           *string
	CreatorRequestID       *string
	DestinationArn         *string
	ID                     *string
	Name                   *string
	OwnerID                *string
	ShareStatus            r53rtypes.ShareStatus
	Status                 r53rtypes.ResolverQueryLogConfigStatus
}

func (r *Route53ResolverQueryLogConfig) Remove(ctx context.Context) error {
	var notFound *r53rtypes.ResourceNotFoundException

	// disassociate resources (VPCs)
	for _, resourceID := range r.resourceAssociationIds {
		_, err := r.svc.DisassociateResolverQueryLogConfig(ctx, &r53r.DisassociateResolverQueryLogConfigInput{
			ResolverQueryLogConfigId: r.ID,
			ResourceId:               resourceID,
		})

		if err != nil {
			// ignore, resource has probably been disassociated
			if errors.As(err, &notFound) {
				continue
			}
			return err
		}
	}

	// Delete QLC
	_, err := r.svc.DeleteResolverQueryLogConfig(ctx, &r53r.DeleteResolverQueryLogConfigInput{
		ResolverQueryLogConfigId: r.ID,
	})

	return err
}

func (r *Route53ResolverQueryLogConfig) Properties() types.Properties {
	props := types.NewPropertiesFromStruct(r)
	// TODO(v4): remove backward-compat properties
	props.Set("Id", r.ID)
	props.Set("CreatorRequestId", r.CreatorRequestID)
	props.Set("OwnerId", r.OwnerID)
	return props
}

func (r *Route53ResolverQueryLogConfig) String() string {
	return *r.ID
}

// qlcsToAssociationIds - Associate all the QLC resource ids to their query log config ID to be
// disassociated before deleting the QLC.
func qlcsToAssociationIds(ctx context.Context, svc Route53ResolverAPI) (map[string][]*string, error) {
	resourceAssociations := map[string][]*string{}

	params := &r53r.ListResolverQueryLogConfigAssociationsInput{}

	for {
		resp, err := svc.ListResolverQueryLogConfigAssociations(ctx, params)

		if err != nil {
			return nil, err
		}

		for _, qlcAssociation := range resp.ResolverQueryLogConfigAssociations {
			resourceID := qlcAssociation.ResourceId
			if resourceID != nil {
				qlcID := *qlcAssociation.ResolverQueryLogConfigId

				if _, ok := resourceAssociations[qlcID]; !ok {
					resourceAssociations[qlcID] = []*string{resourceID}
				} else {
					resourceAssociations[qlcID] = append(resourceAssociations[qlcID], resourceID)
				}
			}
		}

		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return resourceAssociations, nil
}
