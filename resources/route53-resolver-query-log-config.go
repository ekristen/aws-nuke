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
	var resources []resource.Resource

	if l.svc == nil {
		opts := o.(*nuke.ListerOpts)
		svc := r53r.NewFromConfig(*opts.Config)
		l.svc = svc
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
				CreatorRequestId:       qlc.CreatorRequestId,
				DestinationArn:         qlc.DestinationArn,
				Id:                     qlc.Id,
				OwnerId:                qlc.OwnerId,
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
	CreatorRequestId       *string
	DestinationArn         *string
	Id                     *string
	Name                   *string
	OwnerId                *string
	ShareStatus            r53rtypes.ShareStatus
	Status                 r53rtypes.ResolverQueryLogConfigStatus
}

func (r *Route53ResolverQueryLogConfig) Remove(ctx context.Context) error {
	var notFound *r53rtypes.ResourceNotFoundException

	// disassociate resources (VPCs)
	for _, resourceId := range r.resourceAssociationIds {
		_, err := r.svc.DisassociateResolverQueryLogConfig(ctx, &r53r.DisassociateResolverQueryLogConfigInput{
			ResolverQueryLogConfigId: r.Id,
			ResourceId:               resourceId,
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
		ResolverQueryLogConfigId: r.Id,
	})

	return err
}

func (r *Route53ResolverQueryLogConfig) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
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
			resourceId := qlcAssociation.ResourceId
			if resourceId != nil {
				qlcId := *qlcAssociation.ResolverQueryLogConfigId

				if _, ok := resourceAssociations[qlcId]; !ok {
					resourceAssociations[qlcId] = []*string{resourceId}
				} else {
					resourceAssociations[qlcId] = append(resourceAssociations[qlcId], resourceId)
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
