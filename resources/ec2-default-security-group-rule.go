package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const EC2DefaultSecurityGroupRuleResource = "EC2DefaultSecurityGroupRule"

func init() {
	registry.Register(&registry.Registration{
		Name:   EC2DefaultSecurityGroupRuleResource,
		Scope:  nuke.Account,
		Lister: &EC2DefaultSecurityGroupRuleLister{},
	})
}

type EC2DefaultSecurityGroupRuleLister struct{}

func (l *EC2DefaultSecurityGroupRuleLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := ec2.New(opts.Session)
	resources := make([]resource.Resource, 0)

	sgFilters := []*ec2.Filter{
		{
			Name: aws.String("group-name"),
			Values: []*string{
				aws.String("default"),
			},
		},
	}
	groupIds := make([]*string, 0)
	sgParams := &ec2.DescribeSecurityGroupsInput{Filters: sgFilters}
	err := svc.DescribeSecurityGroupsPages(sgParams,
		func(page *ec2.DescribeSecurityGroupsOutput, lastPage bool) bool {
			for _, group := range page.SecurityGroups {
				groupIds = append(groupIds, group.GroupId)
			}
			return !lastPage
		})
	if err != nil {
		return nil, err
	}

	if len(groupIds) == 0 {
		return resources, nil
	}

	sgRuleFilters := []*ec2.Filter{
		{
			Name:   aws.String("group-id"),
			Values: groupIds,
		},
	}
	sgRuleParams := &ec2.DescribeSecurityGroupRulesInput{Filters: sgRuleFilters}
	err = svc.DescribeSecurityGroupRulesPages(sgRuleParams,
		func(page *ec2.DescribeSecurityGroupRulesOutput, lastPage bool) bool {
			for _, rule := range page.SecurityGroupRules {
				resources = append(resources, &EC2DefaultSecurityGroupRule{
					svc:      svc,
					id:       rule.SecurityGroupRuleId,
					groupID:  rule.GroupId,
					isEgress: rule.IsEgress,
				})
			}
			return !lastPage
		})
	if err != nil {
		return nil, err
	}

	return resources, nil
}

type EC2DefaultSecurityGroupRule struct {
	svc      *ec2.EC2
	id       *string
	groupID  *string
	isEgress *bool
}

func (r *EC2DefaultSecurityGroupRule) Remove(_ context.Context) error {
	rules := make([]*string, 1)
	rules[0] = r.id
	if *r.isEgress {
		params := &ec2.RevokeSecurityGroupEgressInput{
			GroupId:              r.groupID,
			SecurityGroupRuleIds: rules,
		}
		_, err := r.svc.RevokeSecurityGroupEgress(params)

		if err != nil {
			return err
		}
	} else {
		params := &ec2.RevokeSecurityGroupIngressInput{
			GroupId:              r.groupID,
			SecurityGroupRuleIds: rules,
		}
		_, err := r.svc.RevokeSecurityGroupIngress(params)

		if err != nil {
			return err
		}
	}

	return nil
}

func (r *EC2DefaultSecurityGroupRule) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("SecurityGroupId", r.groupID)
	properties.Set("DefaultVPC", true)
	return properties
}

func (r *EC2DefaultSecurityGroupRule) String() string {
	return *r.id
}
