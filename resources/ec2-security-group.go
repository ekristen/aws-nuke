package resources

import (
	"context"
	"errors"
	"fmt"

	"github.com/gotidy/ptr"

	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
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
	svc       *ec2.EC2
	accountID *string
	ingress   []*ec2.IpPermission
	egress    []*ec2.IpPermission

	ID      *string `description:"The ID of the security group."`
	Name    *string `description:"The name of the security group."`
	OwnerID *string `description:"The ID of the AWS account that owns the security group."`
	Tags    []*ec2.Tag
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
					svc:       svc,
					accountID: opts.AccountID,
					ingress:   group.IpPermissions,
					egress:    group.IpPermissionsEgress,
					ID:        group.GroupId,
					Name:      group.GroupName,
					OwnerID:   group.OwnerId,
					Tags:      group.Tags,
				})
			}
			return !lastPage
		}); err != nil {
		return nil, err
	}

	return resources, nil
}

func (r *EC2SecurityGroup) Filter() error {
	if ptr.ToString(r.Name) == "default" {
		return fmt.Errorf("cannot delete group 'default'")
	}

	if ptr.ToString(r.OwnerID) != ptr.ToString(r.accountID) {
		return errors.New("not owned by account, likely shared")
	}

	return nil
}

func (r *EC2SecurityGroup) Remove(_ context.Context) error {
	if len(r.egress) > 0 {
		egressParams := &ec2.RevokeSecurityGroupEgressInput{
			GroupId:       r.ID,
			IpPermissions: r.egress,
		}

		_, _ = r.svc.RevokeSecurityGroupEgress(egressParams)
	}

	if len(r.ingress) > 0 {
		ingressParams := &ec2.RevokeSecurityGroupIngressInput{
			GroupId:       r.ID,
			IpPermissions: r.ingress,
		}

		_, _ = r.svc.RevokeSecurityGroupIngress(ingressParams)
	}

	params := &ec2.DeleteSecurityGroupInput{
		GroupId: r.ID,
	}

	_, err := r.svc.DeleteSecurityGroup(params)
	if err != nil {
		return err
	}

	return nil
}

func (r *EC2SecurityGroup) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *EC2SecurityGroup) String() string {
	return ptr.ToString(r.ID)
}
