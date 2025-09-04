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

const ShieldProtectionResource = "ShieldProtection"

func init() {
	registry.Register(&registry.Registration{
		Name:     ShieldProtectionResource,
		Scope:    nuke.Account,
		Resource: &ShieldProtection{},
		Lister:   &ShieldProtectionLister{},
	})
}

type ShieldProtectionLister struct{}

func (l *ShieldProtectionLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := shield.NewFromConfig(*opts.Config)

	params := &shield.ListProtectionsInput{}
	resources := make([]resource.Resource, 0)

	for {
		resp, err := svc.ListProtections(ctx, params)
		if err != nil {
			return nil, err
		}

		for i := range resp.Protections {
			protection := &resp.Protections[i]

			tags, err := svc.ListTagsForResource(ctx, &shield.ListTagsForResourceInput{
				ResourceARN: protection.ProtectionArn,
			})
			if err != nil {
				return nil, err
			}

			resources = append(resources, &ShieldProtection{
				svc:           svc,
				ID:            protection.Id,
				Name:          protection.Name,
				ResourceArn:   protection.ResourceArn,
				ProtectionArn: protection.ProtectionArn,
				Tags:          &tags.Tags,
			})
		}

		if resp.NextToken == nil {
			break
		}
		params.NextToken = resp.NextToken
	}

	return resources, nil
}

type ShieldProtection struct {
	svc           *shield.Client
	ID            *string            `description:"The unique identifier of the Shield protection"`
	Name          *string            `description:"The name of the Shield protection"`
	ResourceArn   *string            `description:"The ARN of the AWS resource being protected"`
	ProtectionArn *string            `description:"The ARN of the Shield protection"`
	Tags          *[]shieldtypes.Tag `description:"The tags associated with the Shield protection"`
}

func (r *ShieldProtection) Remove(ctx context.Context) error {
	params := &shield.DeleteProtectionInput{
		ProtectionId: r.ID,
	}

	_, err := r.svc.DeleteProtection(ctx, params)
	return err
}

func (r *ShieldProtection) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *ShieldProtection) String() string {
	return *r.ID
}
