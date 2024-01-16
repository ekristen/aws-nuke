package resources

import (
	"context"

	"fmt"

	"github.com/gotidy/ptr"

	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/ekristen/aws-nuke/pkg/nuke"

	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"
)

const EC2SecurityGroupResource = "EC2SecurityGroup"

func init() {
	resource.Register(resource.Registration{
		Name:   EC2SecurityGroupResource,
		Scope:  nuke.Account,
		Lister: &EC2SecurityGroupLister{},
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

func (sg *EC2SecurityGroup) Filter() error {
	if ptr.ToString(sg.name) == "default" {
		return fmt.Errorf("cannot delete group 'default'")
	}

	return nil
}

func (sg *EC2SecurityGroup) Remove(_ context.Context) error {
	if len(sg.egress) > 0 {
		egressParams := &ec2.RevokeSecurityGroupEgressInput{
			GroupId:       sg.id,
			IpPermissions: sg.egress,
		}

		_, _ = sg.svc.RevokeSecurityGroupEgress(egressParams)
	}

	if len(sg.ingress) > 0 {
		ingressParams := &ec2.RevokeSecurityGroupIngressInput{
			GroupId:       sg.id,
			IpPermissions: sg.ingress,
		}

		_, _ = sg.svc.RevokeSecurityGroupIngress(ingressParams)
	}

	params := &ec2.DeleteSecurityGroupInput{
		GroupId: sg.id,
	}

	_, err := sg.svc.DeleteSecurityGroup(params)
	if err != nil {
		return err
	}

	return nil
}

func (sg *EC2SecurityGroup) Properties() types.Properties {
	properties := types.NewProperties()
	for _, tagValue := range sg.group.Tags {
		properties.SetTag(tagValue.Key, tagValue.Value)
	}
	properties.Set("Name", sg.name)
	properties.Set("OwnerID", sg.ownerID)
	return properties
}

func (sg *EC2SecurityGroup) String() string {
	return ptr.ToString(sg.id)
}
