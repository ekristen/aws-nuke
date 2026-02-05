package resources

import (
	"context"
	"testing"

	r53r "github.com/aws/aws-sdk-go-v2/service/route53resolver"
	r53rtypes "github.com/aws/aws-sdk-go-v2/service/route53resolver/types"

	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/v3/mocks/mock_route53resolverv2"
)

func Test_Mock_Route53ResolverFirewallRuleGroup_List(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRoute53Resolver := mock_route53resolverv2.NewMockRoute53ResolverAPI(ctrl)

	expectedFirewallRule1a := &Route53ResolverFirewallRule{ptr.String("test-rule"),
		ptr.String("vpc-123"), ptr.String("A"), nil}
	expectedFirewallRule2a := &Route53ResolverFirewallRule{ptr.String("block-malicious"),
		ptr.String("vpc-123"), nil, nil}
	expectedFirewallRule2b := &Route53ResolverFirewallRule{ptr.String("block-malicious2"), nil, nil, nil}

	paramsFrg1 := &r53r.ListFirewallRulesInput{
		FirewallRuleGroupId: ptr.String("rslvr-frg-1"),
	}

	mockRoute53Resolver.EXPECT().ListFirewallRules(gomock.Any(), paramsFrg1).Return(&r53r.ListFirewallRulesOutput{
		FirewallRules: []r53rtypes.FirewallRule{
			{
				Action:               "ALLOW",
				BlockResponse:        "NODATA",
				CreationTime:         ptr.String("2023-01-01T00:00:00Z"),
				CreatorRequestId:     ptr.String("test-request-1"),
				FirewallDomainListId: expectedFirewallRule1a.FirewallDomainListId,
				FirewallRuleGroupId:  ptr.String("rslvr-frg-1"),
				ModificationTime:     ptr.String("2023-01-01T00:00:00Z"),
				Name:                 expectedFirewallRule1a.Name,
				Qtype:                expectedFirewallRule1a.Qtype,
				Priority:             ptr.Int32(100),
			},
		},
	}, nil)

	paramsFrg2 := &r53r.ListFirewallRulesInput{
		FirewallRuleGroupId: ptr.String("rslvr-frg-2"),
	}

	mockRoute53Resolver.EXPECT().ListFirewallRules(gomock.Any(), paramsFrg2).Return(&r53r.ListFirewallRulesOutput{
		FirewallRules: []r53rtypes.FirewallRule{
			{
				Action:               "BLOCK",
				BlockResponse:        "NXDOMAIN",
				CreationTime:         ptr.String("2023-01-01T00:00:00Z"),
				CreatorRequestId:     ptr.String("test-request-2"),
				FirewallDomainListId: expectedFirewallRule2a.FirewallDomainListId,
				FirewallRuleGroupId:  ptr.String("rslvr-frg-2"),
				ModificationTime:     ptr.String("2023-01-01T00:00:00Z"),
				Name:                 expectedFirewallRule2a.Name,
				Priority:             ptr.Int32(200),
			},
			{
				Action:               "BLOCK",
				BlockResponse:        "NXDOMAIN",
				CreationTime:         ptr.String("2023-01-01T00:00:00Z"),
				CreatorRequestId:     ptr.String("test-request-2"),
				FirewallDomainListId: expectedFirewallRule2b.FirewallDomainListId,
				FirewallRuleGroupId:  ptr.String("rslvr-frg-2"),
				ModificationTime:     ptr.String("2023-01-01T00:00:00Z"),
				Name:                 expectedFirewallRule2b.Name,
				Priority:             ptr.Int32(200),
			},
		},
	}, nil)

	paramsFrg3 := &r53r.ListFirewallRulesInput{
		FirewallRuleGroupId: ptr.String("rslvr-frg-3"),
	}

	mockRoute53Resolver.EXPECT().ListFirewallRules(gomock.Any(), paramsFrg3).Return(&r53r.ListFirewallRulesOutput{
		FirewallRules: []r53rtypes.FirewallRule{},
	}, nil)

	mockRoute53Resolver.EXPECT().ListFirewallRuleGroupAssociations(gomock.Any(),
		gomock.Any()).Return(&r53r.ListFirewallRuleGroupAssociationsOutput{
		FirewallRuleGroupAssociations: []r53rtypes.FirewallRuleGroupAssociation{
			{
				Id:                  ptr.String("rslvr-frgassoc-1"),
				FirewallRuleGroupId: ptr.String("rslvr-frg-1"),
				ManagedOwnerName:    ptr.String("Route 53 Resolver DNS Firewall"),
				Name:                ptr.String("association-1"),
				Priority:            ptr.Int32(100),
				Status:              "COMPLETE",
				StatusMessage:       ptr.String(""),
				VpcId:               ptr.String("vpc-12345"),
			},
			{
				Id:                  ptr.String("rslvr-frgassoc-2"),
				FirewallRuleGroupId: ptr.String("rslvr-frg-2"),
				ManagedOwnerName:    ptr.String("Route 53 Resolver DNS Firewall"),
				Name:                ptr.String("association-2"),
				Priority:            ptr.Int32(200),
				Status:              "COMPLETE",
				StatusMessage:       ptr.String(""),
				VpcId:               ptr.String("vpc-67890"),
			},
		},
	}, nil)

	mockRoute53Resolver.EXPECT().ListFirewallRuleGroups(gomock.Any(), gomock.Any()).Return(&r53r.ListFirewallRuleGroupsOutput{
		FirewallRuleGroups: []r53rtypes.FirewallRuleGroupMetadata{
			{
				Arn:              ptr.String("arn:aws:route53resolver:us-east-1:123456123456:firewall-rule-group/rslvr-frg-1"),
				CreatorRequestId: ptr.String("SomeAwsServiceCommand"),
				Id:               ptr.String("rslvr-frg-1"),
				OwnerId:          ptr.String("Route 53 Resolver DNS Firewall"),
				Name:             ptr.String("frgNum1"),
				ShareStatus:      r53rtypes.ShareStatusNotShared,
			},
			{
				Arn:              ptr.String("arn:aws:route53resolver:us-east-1:123456123456:firewall-rule-group/rslvr-frg-2"),
				CreatorRequestId: ptr.String("SomeAwsServiceCommand"),
				Id:               ptr.String("rslvr-frg-2"),
				OwnerId:          ptr.String("Route 53 Resolver DNS Firewall"),
				Name:             ptr.String("frgNum2"),
				ShareStatus:      r53rtypes.ShareStatusSharedByMe,
			},
			{
				Arn:              ptr.String("arn:aws:route53resolver:us-east-1:123456123456:firewall-rule-group/rslvr-frg-3"),
				CreatorRequestId: ptr.String("AWSConsole.88.1762108398672"),
				Id:               ptr.String("rslvr-frg-3"),
				OwnerId:          ptr.String("SomeOwnerId"),
				Name:             ptr.String("UserCreatedRuleGroup"),
				ShareStatus:      r53rtypes.ShareStatusSharedWithMe,
			},
		},
	}, nil)

	lister := &Route53ResolverFirewallRuleGroupLister{
		svc: mockRoute53Resolver,
	}

	resources, err := lister.List(context.TODO(), testListerOpts)
	a.Nil(err)
	a.Len(resources, 3)

	expectedResources := []resource.Resource{
		&Route53ResolverFirewallRuleGroup{
			svc:               mockRoute53Resolver,
			vpcAssociationIds: []*string{ptr.String("rslvr-frgassoc-1")},
			rules:             []*Route53ResolverFirewallRule{expectedFirewallRule1a},
			Arn:               ptr.String("arn:aws:route53resolver:us-east-1:123456123456:firewall-rule-group/rslvr-frg-1"),
			CreatorRequestId:  ptr.String("SomeAwsServiceCommand"),
			Id:                ptr.String("rslvr-frg-1"),
			OwnerId:           ptr.String("Route 53 Resolver DNS Firewall"),
			Name:              ptr.String("frgNum1"),
			ShareStatus:       r53rtypes.ShareStatusNotShared,
		},
		&Route53ResolverFirewallRuleGroup{
			svc:               mockRoute53Resolver,
			vpcAssociationIds: []*string{ptr.String("rslvr-frgassoc-2")},
			rules:             []*Route53ResolverFirewallRule{expectedFirewallRule2a, expectedFirewallRule2b},
			Arn:               ptr.String("arn:aws:route53resolver:us-east-1:123456123456:firewall-rule-group/rslvr-frg-2"),
			CreatorRequestId:  ptr.String("SomeAwsServiceCommand"),
			Id:                ptr.String("rslvr-frg-2"),
			OwnerId:           ptr.String("Route 53 Resolver DNS Firewall"),
			Name:              ptr.String("frgNum2"),
			ShareStatus:       r53rtypes.ShareStatusSharedByMe,
		},
		&Route53ResolverFirewallRuleGroup{
			svc:               mockRoute53Resolver,
			vpcAssociationIds: nil,
			rules:             []*Route53ResolverFirewallRule{},
			Arn:               ptr.String("arn:aws:route53resolver:us-east-1:123456123456:firewall-rule-group/rslvr-frg-3"),
			CreatorRequestId:  ptr.String("AWSConsole.88.1762108398672"),
			Id:                ptr.String("rslvr-frg-3"),
			OwnerId:           ptr.String("SomeOwnerId"),
			Name:              ptr.String("UserCreatedRuleGroup"),
			ShareStatus:       r53rtypes.ShareStatusSharedWithMe,
		},
	}

	a.Equal(expectedResources, resources)
}

func Test_Mock_Route53ResolverFirewallRuleGroup_Remove(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRoute53Resolver := mock_route53resolverv2.NewMockRoute53ResolverAPI(ctrl)

	mockRoute53Resolver.EXPECT().
		DisassociateFirewallRuleGroup(gomock.Any(), gomock.Any()).
		Return(&r53r.DisassociateFirewallRuleGroupOutput{}, nil)

	mockRoute53Resolver.EXPECT().
		DeleteFirewallRule(gomock.Any(), gomock.Any()).
		Return(&r53r.DeleteFirewallRuleOutput{}, nil)

	mockRoute53Resolver.EXPECT().
		DeleteFirewallRuleGroup(gomock.Any(), gomock.Any()).
		Return(&r53r.DeleteFirewallRuleGroupOutput{}, nil)

	firewall_rules := []*Route53ResolverFirewallRule{
		{
			Name:                       ptr.String("rule1"),
			Qtype:                      ptr.String("AAAA"),
			FirewallThreatProtectionId: ptr.String("ftpid_123"),
		},
	}

	frg := &Route53ResolverFirewallRuleGroup{
		svc:               mockRoute53Resolver,
		vpcAssociationIds: []*string{ptr.String("rslvr-frgassoc-1")},
		rules:             firewall_rules,
		Arn:               ptr.String("arn:aws:route53resolver:us-east-1:123456123456:firewall-rule-group/rslvr-frg-3"),
		CreatorRequestId:  ptr.String("SomeAwsServiceCommand"),
		Id:                ptr.String("rslvr-frg-3"),
		OwnerId:           ptr.String("Route 53 Resolver DNS Firewall"),
		Name:              ptr.String("Internet Resolver"),
	}

	err := frg.Remove(context.TODO())
	a.Nil(err)
}

func Test_Mock_Route53ResolverFirewallRuleGroup_Properties(t *testing.T) {
	a := assert.New(t)

	frg := &Route53ResolverFirewallRuleGroup{
		Arn:              ptr.String("arn:aws:route53resolver:us-east-1:123456123456:firewall-rule-group/rslvr-frg-3"),
		CreatorRequestId: ptr.String("SomeAwsServiceCommand"),
		Id:               ptr.String("rslvr-frg-3"),
		OwnerId:          ptr.String("Route 53 Resolver DNS Firewall"),
		Name:             ptr.String("Internet Resolver"),
	}

	a.NotNil(frg.Properties())
}
