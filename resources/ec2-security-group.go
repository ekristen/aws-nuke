package resources

import (
	"context"
	"fmt"

	"github.com/gotidy/ptr"

	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"
)

const EC2SecurityGroupResource = "EC2SecurityGroup"

func init() {
	registry.Register(&registry.Registration{
		Name:     EC2SecurityGroupResource,
		Scope:    nuke.Account,
		Resource: &EC2SecurityGroup{},
		Lister:   &EC2SecurityGroupLister{},
		DependsOn: []string{
			ELBv2Resource,
			EC2DefaultSecurityGroupRuleResource,
		},
	})
}

type EC2SecurityGroup struct {
	svc     *ec2.EC2
	group   *ec2.SecurityGroup
	id      *string
	name    *string
	ingress []*ec2.IpPermission
	egress  []*ec2.IpPermission
	ownerID *string
}

type EC2SecurityGroupLister struct{}

func (l *EC2SecurityGroupLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := ec2.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &ec2.DescribeSecurityGroupsInput{}
	if err := svc.DescribeSecurityGroupsPages(params,
		func(page *ec2.DescribeSecurityGroupsOutput, lastPage bool) bool {
			for _, group := range page.SecurityGroups {
				resources = append(resources, &EC2SecurityGroup{
					svc:     svc,
					group:   group,
					id:      group.GroupId,
					name:    group.GroupName,
					ingress: group.IpPermissions,
					egress:  group.IpPermissionsEgress,
					ownerID: group.OwnerId,
				})
			}
			return !lastPage
		}); err != nil {
		return nil, err
	}

	return resources, nil
}

func (r *EC2SecurityGroup) Filter() error {
	if ptr.ToString(r.name) == "default" {
		return fmt.Errorf("cannot delete group 'default'")
	}

	return nil
}

func (r *EC2SecurityGroup) Remove(_ context.Context) error {
	if len(r.egress) > 0 {
		egressParams := &ec2.RevokeSecurityGroupEgressInput{
			GroupId:       r.id,
			IpPermissions: r.egress,
		}

		_, _ = r.svc.RevokeSecurityGroupEgress(egressParams)
	}

	if len(r.ingress) > 0 {
		ingressParams := &ec2.RevokeSecurityGroupIngressInput{
			GroupId:       r.id,
			IpPermissions: r.ingress,
		}

		_, _ = r.svc.RevokeSecurityGroupIngress(ingressParams)
	}

	params := &ec2.DeleteSecurityGroupInput{
		GroupId: r.id,
	}

	_, err := r.svc.DeleteSecurityGroup(params)
	if err != nil {
		return err
	}

	return nil
}

func (r *EC2SecurityGroup) Properties() types.Properties {
	properties := types.NewProperties()
	for _, tagValue := range r.group.Tags {
		properties.SetTag(tagValue.Key, tagValue.Value)
	}
	properties.Set("Name", r.name)
	properties.Set("OwnerID", r.ownerID)
	return properties
}

func (r *EC2SecurityGroup) String() string {
	return ptr.ToString(r.id)
}
