package resources

import (
	"context"
	"testing"

	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	r53r "github.com/aws/aws-sdk-go-v2/service/route53resolver"
	r53rtypes "github.com/aws/aws-sdk-go-v2/service/route53resolver/types"

	liberrors "github.com/ekristen/libnuke/pkg/errors"
	"github.com/ekristen/libnuke/pkg/resource"
	libsettings "github.com/ekristen/libnuke/pkg/settings"

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
				FirewallDomainListId: expectedFirewallRule1a.FirewallDomainListID,
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
				FirewallDomainListId: expectedFirewallRule2a.FirewallDomainListID,
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
				FirewallDomainListId: expectedFirewallRule2b.FirewallDomainListID,
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
				MutationProtection:  r53rtypes.MutationProtectionStatusDisabled,
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
				MutationProtection:  r53rtypes.MutationProtectionStatusEnabled,
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

	expectedVpcAssoc1 := &Route53ResolverFirewallRuleGroupVpcAssociation{ptr.String("rslvr-frgassoc-1"),
		r53rtypes.MutationProtectionStatusDisabled}
	expectedVpcAssoc2 := &Route53ResolverFirewallRuleGroupVpcAssociation{ptr.String("rslvr-frgassoc-2"),
		r53rtypes.MutationProtectionStatusEnabled}

	expectedResources := []resource.Resource{
		&Route53ResolverFirewallRuleGroup{
			svc:              mockRoute53Resolver,
			vpcAssociations:  []*Route53ResolverFirewallRuleGroupVpcAssociation{expectedVpcAssoc1},
			rules:            []*Route53ResolverFirewallRule{expectedFirewallRule1a},
			Arn:              ptr.String("arn:aws:route53resolver:us-east-1:123456123456:firewall-rule-group/rslvr-frg-1"),
			CreatorRequestID: ptr.String("SomeAwsServiceCommand"),
			ID:               ptr.String("rslvr-frg-1"),
			OwnerID:          ptr.String("Route 53 Resolver DNS Firewall"),
			Name:             ptr.String("frgNum1"),
			ShareStatus:      r53rtypes.ShareStatusNotShared,
		},
		&Route53ResolverFirewallRuleGroup{
			svc:              mockRoute53Resolver,
			vpcAssociations:  []*Route53ResolverFirewallRuleGroupVpcAssociation{expectedVpcAssoc2},
			rules:            []*Route53ResolverFirewallRule{expectedFirewallRule2a, expectedFirewallRule2b},
			Arn:              ptr.String("arn:aws:route53resolver:us-east-1:123456123456:firewall-rule-group/rslvr-frg-2"),
			CreatorRequestID: ptr.String("SomeAwsServiceCommand"),
			ID:               ptr.String("rslvr-frg-2"),
			OwnerID:          ptr.String("Route 53 Resolver DNS Firewall"),
			Name:             ptr.String("frgNum2"),
			ShareStatus:      r53rtypes.ShareStatusSharedByMe,
		},
		&Route53ResolverFirewallRuleGroup{
			svc:              mockRoute53Resolver,
			vpcAssociations:  nil,
			rules:            []*Route53ResolverFirewallRule{},
			Arn:              ptr.String("arn:aws:route53resolver:us-east-1:123456123456:firewall-rule-group/rslvr-frg-3"),
			CreatorRequestID: ptr.String("AWSConsole.88.1762108398672"),
			ID:               ptr.String("rslvr-frg-3"),
			OwnerID:          ptr.String("SomeOwnerId"),
			Name:             ptr.String("UserCreatedRuleGroup"),
			ShareStatus:      r53rtypes.ShareStatusSharedWithMe,
		},
	}

	a.Equal(expectedResources, resources)
}

func Test_Mock_Route53ResolverFirewallRuleGroup_Remove(t *testing.T) {
	isEnabled := true
	notEnabled := false

	mpTestConfigs := []struct {
		name       string
		mpDisabled *bool
	}{
		{"delete_protection_enabled", &isEnabled},
		{"delete_protection_disabled", &notEnabled},
		{"delete_protection_unset", nil},
	}

	for _, tc := range mpTestConfigs {
		t.Run(tc.name, func(t *testing.T) {
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

			firewallRules := []*Route53ResolverFirewallRule{
				{
					Name:                       ptr.String("rule1"),
					Qtype:                      ptr.String("AAAA"),
					FirewallThreatProtectionID: ptr.String("ftpid_123"),
				},
			}

			mockRoute53Resolver.EXPECT().ListFirewallRuleGroupAssociations(gomock.Any(),
				gomock.Any()).Return(&r53r.ListFirewallRuleGroupAssociationsOutput{
				FirewallRuleGroupAssociations: []r53rtypes.FirewallRuleGroupAssociation{},
			}, nil)

			expectedVpcAssoc1 := &Route53ResolverFirewallRuleGroupVpcAssociation{ptr.String("rslvr-frgassoc-1"),
				r53rtypes.MutationProtectionStatusEnabled}

			if tc.mpDisabled != nil && *tc.mpDisabled {
				// DisableDeletionProtection set in this case, expect mutation protection to be disabled for association #2
				// which has mutation protection enabled
				mockRoute53Resolver.EXPECT().
					UpdateFirewallRuleGroupAssociation(context.TODO(), gomock.Eq(&r53r.UpdateFirewallRuleGroupAssociationInput{
						FirewallRuleGroupAssociationId: ptr.String("rslvr-frgassoc-1"),
						MutationProtection:             r53rtypes.MutationProtectionStatusDisabled,
					})).Return(&r53r.UpdateFirewallRuleGroupAssociationOutput{}, nil)
			} else {
				// DisableDeletionProtection NOT set in this case so we shouldn't disable mutation protection
				mockRoute53Resolver.EXPECT().UpdateFirewallRuleGroupAssociation(gomock.Any(), gomock.Any()).Times(0)
			}

			mpSettings := &libsettings.Setting{}
			if tc.mpDisabled != nil {
				if *tc.mpDisabled {
					mpSettings = &libsettings.Setting{
						"DisableDeletionProtection": true,
					}
				} else {
					mpSettings = &libsettings.Setting{
						"DisableDeletionProtection": false,
					}
				}
			}

			frg := &Route53ResolverFirewallRuleGroup{
				settings:         mpSettings,
				svc:              mockRoute53Resolver,
				vpcAssociations:  []*Route53ResolverFirewallRuleGroupVpcAssociation{expectedVpcAssoc1},
				rules:            firewallRules,
				Arn:              ptr.String("arn:aws:route53resolver:us-east-1:123456123456:firewall-rule-group/rslvr-frg-3"),
				CreatorRequestID: ptr.String("SomeAwsServiceCommand"),
				ID:               ptr.String("rslvr-frg-3"),
				OwnerID:          ptr.String("Route 53 Resolver DNS Firewall"),
				Name:             ptr.String("Internet Resolver"),
			}

			err := frg.Remove(context.TODO())
			a.Nil(err)
		})
	}
}

func Test_Mock_Route53ResolverFirewallRuleGroup_Properties(t *testing.T) {
	a := assert.New(t)

	frg := &Route53ResolverFirewallRuleGroup{
		Arn:              ptr.String("arn:aws:route53resolver:us-east-1:123456123456:firewall-rule-group/rslvr-frg-3"),
		CreatorRequestID: ptr.String("SomeAwsServiceCommand"),
		ID:               ptr.String("rslvr-frg-3"),
		OwnerID:          ptr.String("Route 53 Resolver DNS Firewall"),
		Name:             ptr.String("Internet Resolver"),
	}

	a.NotNil(frg.Properties())
}

func Test_Mock_Route53ResolverFirewallRuleGroup_waitForAssociationsToStabilize(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRoute53Resolver := mock_route53resolverv2.NewMockRoute53ResolverAPI(ctrl)

	testCases := []struct {
		name   string
		status *string
	}{
		{"association deleting", ptr.String(string(r53rtypes.FirewallRuleGroupAssociationStatusDeleting))},
		{"association complete", ptr.String(string(r53rtypes.FirewallRuleGroupAssociationStatusComplete))},
		{"no associations", nil},
	}

	frg := &Route53ResolverFirewallRuleGroup{
		settings:         &libsettings.Setting{},
		svc:              mockRoute53Resolver,
		vpcAssociations:  []*Route53ResolverFirewallRuleGroupVpcAssociation{},
		rules:            []*Route53ResolverFirewallRule{},
		Arn:              ptr.String("arn:aws:route53resolver:us-east-1:123456123456:firewall-rule-group/rslvr-frg-3"),
		CreatorRequestID: ptr.String("SomeAwsServiceCommand"),
		ID:               ptr.String("rslvr-frg-3"),
		OwnerID:          ptr.String("Route 53 Resolver DNS Firewall"),
		Name:             ptr.String("Internet Resolver"),
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.status != nil {
				mockRoute53Resolver.EXPECT().ListFirewallRuleGroupAssociations(gomock.Any(),
					gomock.Any()).Return(&r53r.ListFirewallRuleGroupAssociationsOutput{
					FirewallRuleGroupAssociations: []r53rtypes.FirewallRuleGroupAssociation{
						{
							Id:                  ptr.String("rslvr-frgassoc-1"),
							FirewallRuleGroupId: ptr.String("rslvr-frg-1"),
							ManagedOwnerName:    ptr.String("Route 53 Resolver DNS Firewall"),
							Name:                ptr.String("association-1"),
							Priority:            ptr.Int32(100),
							Status:              r53rtypes.FirewallRuleGroupAssociationStatus(*tc.status),
							MutationProtection:  r53rtypes.MutationProtectionStatusDisabled,
							StatusMessage:       ptr.String(""),
							VpcId:               ptr.String("vpc-12345"),
						},
					},
				}, nil)
			} else {
				mockRoute53Resolver.EXPECT().ListFirewallRuleGroupAssociations(gomock.Any(),
					gomock.Any()).Return(&r53r.ListFirewallRuleGroupAssociationsOutput{
					FirewallRuleGroupAssociations: []r53rtypes.FirewallRuleGroupAssociation{},
				}, nil)
			}
		})

		err := waitForAssociationToStabilize(context.TODO(), mockRoute53Resolver, frg)

		if tc.status != nil && *tc.status == (string(r53rtypes.FirewallRuleGroupAssociationStatusDeleting)) {
			var expectedErrType liberrors.ErrHoldResource
			a.NotNil(err)
			a.ErrorAs(err, &expectedErrType)
		} else {
			a.Nil(err)
		}
	}
}
