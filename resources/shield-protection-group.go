package resources

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/shield"
	shieldtypes "github.com/aws/aws-sdk-go-v2/service/shield/types"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const ShieldProtectionGroupResource = "ShieldProtectionGroup"

func init() {
	registry.Register(&registry.Registration{
		Name:     ShieldProtectionGroupResource,
		Scope:    nuke.Account,
		Resource: &ShieldProtectionGroup{},
		Lister:   &ShieldProtectionGroupLister{},
	})
}

type ShieldProtectionGroupLister struct{}

func (l *ShieldProtectionGroupLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := shield.NewFromConfig(*opts.Config)

	params := &shield.ListProtectionGroupsInput{}
	resources := make([]resource.Resource, 0)

	for {
		resp, err := svc.ListProtectionGroups(ctx, params)
		if err != nil {
			return nil, err
		}

		for i := range resp.ProtectionGroups {
			group := &resp.ProtectionGroups[i]

			tags, err := svc.ListTagsForResource(ctx, &shield.ListTagsForResourceInput{
				ResourceARN: group.ProtectionGroupArn,
			})
			if err != nil {
				return nil, err
			}

			resources = append(resources, &ShieldProtectionGroup{
				svc:                svc,
				ProtectionGroupID:  group.ProtectionGroupId,
				Aggregation:        &group.Aggregation,
				Pattern:            &group.Pattern,
				ResourceType:       &group.ResourceType,
				Members:            &group.Members,
				ProtectionGroupArn: group.ProtectionGroupArn,
				Tags:               &tags.Tags,
			})
		}

		if resp.NextToken == nil {
			break
		}
		params.NextToken = resp.NextToken
	}

	return resources, nil
}

type ShieldProtectionGroup struct {
	svc                *shield.Client
	ProtectionGroupID  *string                                 `description:"The unique identifier of the Shield protection group"`
	Aggregation        *shieldtypes.ProtectionGroupAggregation `description:"The aggregation type for the protection group"`
	Pattern            *shieldtypes.ProtectionGroupPattern     `description:"The pattern for the protection group"`
	ResourceType       *shieldtypes.ProtectedResourceType      `description:"The resource type for the protection group"`
	Members            *[]string                               `description:"The list of resource ARNs that are members of the protection group"` //nolint:lll
	ProtectionGroupArn *string                                 `description:"The ARN of the Shield protection group"`
	Tags               *[]shieldtypes.Tag                      `description:"The tags associated with the Shield protection group"`
}

func (r *ShieldProtectionGroup) Remove(ctx context.Context) error {
	params := &shield.DeleteProtectionGroupInput{
		ProtectionGroupId: r.ProtectionGroupID,
	}

	_, err := r.svc.DeleteProtectionGroup(ctx, params)
	return err
}

func (r *ShieldProtectionGroup) Properties() types.Properties {
	props := types.NewPropertiesFromStruct(r)
	// TODO(v4): remove backward-compat property
	props.Set("ProtectionGroupId", r.ProtectionGroupID)
	return props
}

func (r *ShieldProtectionGroup) String() string {
	return *r.ProtectionGroupID
}
