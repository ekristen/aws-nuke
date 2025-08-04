package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/ec2"

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

func (l *EC2VerifiedAccessGroupLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := ec2.New(opts.Session)

	params := &ec2.DescribeVerifiedAccessGroupsInput{}
	resources := make([]resource.Resource, 0)

	for {
		resp, err := svc.DescribeVerifiedAccessGroups(params)
		if err != nil {
			return nil, err
		}

		for _, group := range resp.VerifiedAccessGroups {
			resources = append(resources, &EC2VerifiedAccessGroup{
				svc:   svc,
				group: group,
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
	svc   *ec2.EC2
	group *ec2.VerifiedAccessGroup
}

func (r *EC2VerifiedAccessGroup) Remove(_ context.Context) error {
	params := &ec2.DeleteVerifiedAccessGroupInput{
		VerifiedAccessGroupId: r.group.VerifiedAccessGroupId,
	}

	_, err := r.svc.DeleteVerifiedAccessGroup(params)
	return err
}

func (r *EC2VerifiedAccessGroup) Properties() types.Properties {
	properties := types.NewProperties()

	for _, tag := range r.group.Tags {
		properties.SetTag(tag.Key, tag.Value)
	}

	properties.Set("ID", r.group.VerifiedAccessGroupId)
	properties.Set("InstanceID", r.group.VerifiedAccessInstanceId)
	properties.Set("Description", r.group.Description)
	properties.Set("Owner", r.group.Owner)
	properties.Set("CreationTime", r.group.CreationTime)
	properties.Set("LastUpdatedTime", r.group.LastUpdatedTime)

	return properties
}

func (r *EC2VerifiedAccessGroup) String() string {
	return *r.group.VerifiedAccessGroupId
}
