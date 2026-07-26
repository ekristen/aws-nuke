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
	libsettings "github.com/ekristen/libnuke/pkg/settings"

	"github.com/ekristen/aws-nuke/v3/mocks/mock_route53resolverv2"
)

func Test_Mock_Route53ResolverFirewallRuleGroupAssociation_List(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRoute53Resolver := mock_route53resolverv2.NewMockRoute53ResolverAPI(ctrl)

	mockRoute53Resolver.EXPECT().ListFirewallRuleGroupAssociations(gomock.Any(),
		gomock.Any()).Return(&r53r.ListFirewallRuleGroupAssociationsOutput{
		FirewallRuleGroupAssociations: []r53rtypes.FirewallRuleGroupAssociation{
			{
				Id:                  ptr.String("rslvr-frgassoc-1"),
				Arn:                 ptr.String("arn:aws:route53resolver:us-east-1:123456123456:firewall-rule-group-association/rslvr-frgassoc-1"),
				FirewallRuleGroupId: ptr.String("rslvr-frg-1"),
				Name:                ptr.String("association-1"),
				Priority:            ptr.Int32(100),
				Status:              r53rtypes.FirewallRuleGroupAssociationStatusComplete,
				MutationProtection:  r53rtypes.MutationProtectionStatusDisabled,
				VpcId:               ptr.String("vpc-12345"),
			},
			{
				Id:                  ptr.String("rslvr-frgassoc-2"),
				Arn:                 ptr.String("arn:aws:route53resolver:us-east-1:123456123456:firewall-rule-group-association/rslvr-frgassoc-2"),
				FirewallRuleGroupId: ptr.String("rslvr-frg-2"),
				Name:                ptr.String("association-2"),
				Priority:            ptr.Int32(200),
				Status:              r53rtypes.FirewallRuleGroupAssociationStatusComplete,
				MutationProtection:  r53rtypes.MutationProtectionStatusEnabled,
				VpcId:               ptr.String("vpc-67890"),
			},
		},
	}, nil)

	lister := &Route53ResolverFirewallRuleGroupAssociationLister{
		svc: mockRoute53Resolver,
	}

	resources, err := lister.List(context.TODO(), testListerOpts)
	a.Nil(err)
	a.Len(resources, 2)

	first := resources[0].(*Route53ResolverFirewallRuleGroupAssociation)
	a.Equal("rslvr-frgassoc-1", *first.ID)
	a.Equal("rslvr-frg-1", *first.FirewallRuleGroupID)
	a.Equal("vpc-12345", *first.VpcID)
	a.Equal(r53rtypes.MutationProtectionStatusDisabled, first.MutationProtection)

	second := resources[1].(*Route53ResolverFirewallRuleGroupAssociation)
	a.Equal("rslvr-frgassoc-2", *second.ID)
	a.Equal(r53rtypes.MutationProtectionStatusEnabled, second.MutationProtection)
}

func Test_Mock_Route53ResolverFirewallRuleGroupAssociation_Remove(t *testing.T) {
	isEnabled := true
	notEnabled := false

	testCases := []struct {
		name               string
		disableProtection  *bool
		mutationProtection r53rtypes.MutationProtectionStatus
		expectUpdate       bool
	}{
		{"protection_enabled_setting_on", &isEnabled, r53rtypes.MutationProtectionStatusEnabled, true},
		{"protection_enabled_setting_off", &notEnabled, r53rtypes.MutationProtectionStatusEnabled, false},
		{"protection_enabled_setting_unset", nil, r53rtypes.MutationProtectionStatusEnabled, false},
		{"protection_disabled_setting_on", &isEnabled, r53rtypes.MutationProtectionStatusDisabled, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			a := assert.New(t)
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRoute53Resolver := mock_route53resolverv2.NewMockRoute53ResolverAPI(ctrl)

			if tc.expectUpdate {
				mockRoute53Resolver.EXPECT().
					UpdateFirewallRuleGroupAssociation(gomock.Any(), gomock.Eq(&r53r.UpdateFirewallRuleGroupAssociationInput{
						FirewallRuleGroupAssociationId: ptr.String("rslvr-frgassoc-1"),
						MutationProtection:             r53rtypes.MutationProtectionStatusDisabled,
					})).Return(&r53r.UpdateFirewallRuleGroupAssociationOutput{}, nil)
			} else {
				mockRoute53Resolver.EXPECT().
					UpdateFirewallRuleGroupAssociation(gomock.Any(), gomock.Any()).Times(0)
			}

			mockRoute53Resolver.EXPECT().
				DisassociateFirewallRuleGroup(gomock.Any(), gomock.Eq(&r53r.DisassociateFirewallRuleGroupInput{
					FirewallRuleGroupAssociationId: ptr.String("rslvr-frgassoc-1"),
				})).Return(&r53r.DisassociateFirewallRuleGroupOutput{}, nil)

			settings := &libsettings.Setting{}
			if tc.disableProtection != nil {
				settings = &libsettings.Setting{
					"DisableDeletionProtection": *tc.disableProtection,
				}
			}

			assoc := &Route53ResolverFirewallRuleGroupAssociation{
				svc:                 mockRoute53Resolver,
				settings:            settings,
				ID:                  ptr.String("rslvr-frgassoc-1"),
				FirewallRuleGroupID: ptr.String("rslvr-frg-1"),
				VpcID:               ptr.String("vpc-12345"),
				MutationProtection:  tc.mutationProtection,
			}

			err := assoc.Remove(context.TODO())
			a.Nil(err)
		})
	}
}

func Test_Mock_Route53ResolverFirewallRuleGroupAssociation_Properties(t *testing.T) {
	a := assert.New(t)

	assoc := &Route53ResolverFirewallRuleGroupAssociation{
		ID:                  ptr.String("rslvr-frgassoc-1"),
		Name:                ptr.String("association-1"),
		FirewallRuleGroupID: ptr.String("rslvr-frg-1"),
		VpcID:               ptr.String("vpc-12345"),
		Status:              r53rtypes.FirewallRuleGroupAssociationStatusComplete,
		MutationProtection:  r53rtypes.MutationProtectionStatusEnabled,
	}

	props := assoc.Properties()
	a.Equal("rslvr-frgassoc-1", props.Get("ID"))
	a.Equal("rslvr-frg-1", props.Get("FirewallRuleGroupID"))
	a.Equal("vpc-12345", props.Get("VpcID"))
	a.Equal("COMPLETE", props.Get("Status"))
	a.Equal("rslvr-frgassoc-1", assoc.String())
}

func Test_Mock_Route53ResolverFirewallRuleGroupAssociation_Filter(t *testing.T) {
	a := assert.New(t)

	managed := &Route53ResolverFirewallRuleGroupAssociation{
		ID:               ptr.String("rslvr-frgassoc-1"),
		ManagedOwnerName: ptr.String("Firewall Manager"),
	}
	a.Error(managed.Filter())

	unmanaged := &Route53ResolverFirewallRuleGroupAssociation{
		ID: ptr.String("rslvr-frgassoc-2"),
	}
	a.Nil(unmanaged.Filter())
}

func Test_Mock_Route53ResolverFirewallRuleGroupAssociation_HandleWait(t *testing.T) {
	testCases := []struct {
		name      string
		present   bool
		status    r53rtypes.FirewallRuleGroupAssociationStatus
		expectErr bool
	}{
		{"status_deleting", true, r53rtypes.FirewallRuleGroupAssociationStatusDeleting, true},
		{"status_updating", true, r53rtypes.FirewallRuleGroupAssociationStatusUpdating, true},
		{"status_complete", true, r53rtypes.FirewallRuleGroupAssociationStatusComplete, false},
		{"status_gone", false, "", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			a := assert.New(t)
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRoute53Resolver := mock_route53resolverv2.NewMockRoute53ResolverAPI(ctrl)

			associations := []r53rtypes.FirewallRuleGroupAssociation{}
			if tc.present {
				associations = append(associations, r53rtypes.FirewallRuleGroupAssociation{
					Id:                  ptr.String("rslvr-frgassoc-1"),
					FirewallRuleGroupId: ptr.String("rslvr-frg-1"),
					Status:              tc.status,
				})
			}

			mockRoute53Resolver.EXPECT().ListFirewallRuleGroupAssociations(gomock.Any(),
				gomock.Any()).Return(&r53r.ListFirewallRuleGroupAssociationsOutput{
				FirewallRuleGroupAssociations: associations,
			}, nil)

			assoc := &Route53ResolverFirewallRuleGroupAssociation{
				svc:                 mockRoute53Resolver,
				ID:                  ptr.String("rslvr-frgassoc-1"),
				FirewallRuleGroupID: ptr.String("rslvr-frg-1"),
			}

			err := assoc.HandleWait(context.TODO())

			if tc.expectErr {
				var expectedErrType liberrors.ErrWaitResource
				a.ErrorAs(err, &expectedErrType)
			} else {
				a.Nil(err)
			}
		})
	}
}
