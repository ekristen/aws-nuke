package resources

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const EC2VerifiedAccessGroupResource = "EC2VerifiedAccessGroup"

func init() {
	registry.Register(&registry.Registration{
		Name:     EC2VerifiedAccessGroupResource,
		Scope:    nuke.Account,
		Resource: &EC2VerifiedAccessGroup{},
		Lister:   &EC2VerifiedAccessGroupLister{},
		DependsOn: []string{
			EC2VerifiedAccessEndpointResource,
		},
	})
}

type EC2VerifiedAccessGroupLister struct{}

func (l *EC2VerifiedAccessGroupLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := ec2.NewFromConfig(*opts.Config)

	params := &ec2.DescribeVerifiedAccessGroupsInput{}
	resources := make([]resource.Resource, 0)

	for {
		resp, err := svc.DescribeVerifiedAccessGroups(ctx, params)
		if err != nil {
			return nil, err
		}

		for i := range resp.VerifiedAccessGroups {
			group := &resp.VerifiedAccessGroups[i]
			resources = append(resources, &EC2VerifiedAccessGroup{
				svc:                      svc,
				ID:                       group.VerifiedAccessGroupId,
				Description:              group.Description,
				CreationTime:             group.CreationTime,
				LastUpdatedTime:          group.LastUpdatedTime,
				VerifiedAccessInstanceID: group.VerifiedAccessInstanceId,
				Owner:                    group.Owner,
				Tags:                     group.Tags,
			})
		}

		if resp.NextToken == nil {
			break
		}
		params.NextToken = resp.NextToken
	}

	return resources, nil
}

type EC2VerifiedAccessGroup struct {
	svc                      *ec2.Client
	ID                       *string        `description:"The unique identifier of the Verified Access group"`
	Description              *string        `description:"A description for the Verified Access group"`
	CreationTime             *string        `description:"The timestamp when the Verified Access group was created"`
	LastUpdatedTime          *string        `description:"The timestamp when the Verified Access group was last updated"`
	VerifiedAccessInstanceID *string        `description:"The ID of the Verified Access instance this group belongs to"`
	Owner                    *string        `description:"The AWS account ID that owns the Verified Access group"`
	Tags                     []ec2types.Tag `description:"The tags associated with the Verified Access group"`
}

func (r *EC2VerifiedAccessGroup) Remove(ctx context.Context) error {
	params := &ec2.DeleteVerifiedAccessGroupInput{
		VerifiedAccessGroupId: r.ID,
	}

	_, err := r.svc.DeleteVerifiedAccessGroup(ctx, params)
	return err
}

func (r *EC2VerifiedAccessGroup) Properties() types.Properties {
	props := types.NewPropertiesFromStruct(r)
	props.Set("VerifiedAccessInstanceId", r.VerifiedAccessInstanceID)
	return props
}

func (r *EC2VerifiedAccessGroup) String() string {
	return *r.ID
}
