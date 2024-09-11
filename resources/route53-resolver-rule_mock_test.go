package resources

import (
	"context"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53resolver"

	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/v3/mocks/mock_route53resolveriface"
	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

func Test_Mock_Route53ResolverRule_List(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRoute53Resolver := mock_route53resolveriface.NewMockRoute53ResolverAPI(ctrl)

	mockRoute53Resolver.EXPECT().ListResolverRules(gomock.Any()).Return(&route53resolver.ListResolverRulesOutput{
		ResolverRules: []*route53resolver.ResolverRule{
			{
				Id:         ptr.String("rslvr-rr-1"),
				Name:       ptr.String("rule1"),
				DomainName: ptr.String("example.com"),
			},
			{
				Id:         ptr.String("rslvr-rr-2"),
				Name:       ptr.String("rule2"),
				DomainName: ptr.String("example.org"),
			},
			{
				Id:         ptr.String("rslvr-autodefined-rr-3"),
				Name:       ptr.String("Internet Resolver"),
				DomainName: ptr.String("."),
			},
		},
	}, nil)

	mockRoute53Resolver.EXPECT().ListResolverRuleAssociations(gomock.Any()).Return(&route53resolver.ListResolverRuleAssociationsOutput{
		ResolverRuleAssociations: []*route53resolver.ResolverRuleAssociation{
			{
				ResolverRuleId: ptr.String("rslvr-rr-1"),
				VPCId:          ptr.String("vpc-1"),
			},
			{
				ResolverRuleId: ptr.String("rslvr-rr-2"),
				VPCId:          ptr.String("vpc-2"),
			},
			{
				ResolverRuleId: ptr.String("rslvr-autodefined-rr-3"),
				VPCId:          ptr.String("vpc-3"),
			},
		},
	}, nil)

	lister := &Route53ResolverRuleLister{
		mockSvc: mockRoute53Resolver,
	}

	resources, err := lister.List(context.TODO(), &nuke.ListerOpts{
		Region: &nuke.Region{
			Name: "us-east-2",
		},
		Session: session.Must(session.NewSession()),
	})
	a.Nil(err)
	a.Len(resources, 3)

	expectedResources := []resource.Resource{
		&Route53ResolverRule{
			svc:        mockRoute53Resolver,
			vpcIds:     []*string{ptr.String("vpc-1")},
			ID:         ptr.String("rslvr-rr-1"),
			Name:       ptr.String("rule1"),
			DomainName: ptr.String("example.com"),
		},
		&Route53ResolverRule{
			svc:        mockRoute53Resolver,
			vpcIds:     []*string{ptr.String("vpc-2")},
			ID:         ptr.String("rslvr-rr-2"),
			Name:       ptr.String("rule2"),
			DomainName: ptr.String("example.org"),
		},
		&Route53ResolverRule{
			svc:        mockRoute53Resolver,
			vpcIds:     []*string{ptr.String("vpc-3")},
			ID:         ptr.String("rslvr-autodefined-rr-3"),
			Name:       ptr.String("Internet Resolver"),
			DomainName: ptr.String("."),
		},
	}

	a.Equal(expectedResources, resources)
}

func Test_Mock_Route53ResolverRule_Filter(t *testing.T) {
	a := assert.New(t)

	cases := []struct {
		Name       string
		ID         string
		DomainName string
		Filtered   bool
	}{
		{
			ID:         "rslvr-rr-1",
			DomainName: "example.com",
			Filtered:   false,
		},
		{
			ID:         "rslvr-autodefined-rr-1",
			DomainName: ".",
			Filtered:   true,
		},
	}

	for _, c := range cases {
		name := c.ID
		if c.Filtered {
			name = fmt.Sprintf("filtered/%s", name)
		} else {
			name = fmt.Sprintf("not-filtered/%s", name)
		}

		t.Run(name, func(t *testing.T) {
			rule := &Route53ResolverRule{
				ID:         ptr.String(c.ID),
				DomainName: ptr.String(c.DomainName),
			}

			err := rule.Filter()
			if c.Filtered {
				a.NotNil(err)
			} else {
				a.Nil(err)
			}
		})
	}
}

func Test_Mock_Route53ResolverRule_Remove(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRoute53Resolver := mock_route53resolveriface.NewMockRoute53ResolverAPI(ctrl)

	mockRoute53Resolver.EXPECT().
		DisassociateResolverRule(gomock.Any()).
		Return(&route53resolver.DisassociateResolverRuleOutput{}, nil).Times(2)

	mockRoute53Resolver.EXPECT().
		DeleteResolverRule(gomock.Any()).
		Return(&route53resolver.DeleteResolverRuleOutput{}, nil)

	rule := &Route53ResolverRule{
		svc:        mockRoute53Resolver,
		vpcIds:     []*string{ptr.String("vpc-1"), ptr.String("vpc-2")},
		ID:         ptr.String("rslvr-rr-1"),
		Name:       ptr.String("rule1"),
		DomainName: ptr.String("example.com"),
	}

	err := rule.Remove(context.TODO())
	a.Nil(err)
}

func Test_Mock_Route53ResolverRule_Properties(t *testing.T) {
	a := assert.New(t)

	rule := &Route53ResolverRule{
		ID:         ptr.String("rslvr-rr-1"),
		Name:       ptr.String("rule1"),
		DomainName: ptr.String("example.com"),
	}

	properties := rule.Properties()
	a.Equal("rslvr-rr-1", properties.Get("ID"))
	a.Equal("rule1", properties.Get("Name"))
	a.Equal("example.com", properties.Get("DomainName"))

	a.Equal("rslvr-rr-1 (rule1)", rule.String())

	rule.Name = nil
	a.Equal("rslvr-rr-1 ()", rule.String())
}

func Test_resolverRulesToVpcIDs(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRoute53Resolver := mock_route53resolveriface.NewMockRoute53ResolverAPI(ctrl)

	// Test case: Error from ListResolverRuleAssociations
	mockRoute53Resolver.EXPECT().
		ListResolverRuleAssociations(gomock.Any()).
		Return(nil, aws.ErrMissingEndpoint)

	vpcAssociations, err := resolverRulesToVpcIDs(mockRoute53Resolver)
	a.Nil(vpcAssociations)
	a.NotNil(err)
	a.EqualError(err, aws.ErrMissingEndpoint.Error())

	// Test case: Paginated results with NextToken
	mockRoute53Resolver.EXPECT().
		ListResolverRuleAssociations(&route53resolver.ListResolverRuleAssociationsInput{}).
		Return(&route53resolver.ListResolverRuleAssociationsOutput{
			ResolverRuleAssociations: []*route53resolver.ResolverRuleAssociation{
				{
					ResolverRuleId: ptr.String("rslvr-rr-1"),
					VPCId:          ptr.String("vpc-1"),
				},
			},
			NextToken: ptr.String("token1"),
		}, nil)

	mockRoute53Resolver.EXPECT().
		ListResolverRuleAssociations(&route53resolver.ListResolverRuleAssociationsInput{
			NextToken: ptr.String("token1"),
		}).Return(&route53resolver.ListResolverRuleAssociationsOutput{
		ResolverRuleAssociations: []*route53resolver.ResolverRuleAssociation{
			{
				ResolverRuleId: ptr.String("rslvr-rr-2"),
				VPCId:          ptr.String("vpc-2"),
			},
		},
	}, nil)

	vpcAssociations, err = resolverRulesToVpcIDs(mockRoute53Resolver)
	a.Nil(err)
	a.NotNil(vpcAssociations)
	a.Len(vpcAssociations, 2)
	a.Equal([]*string{ptr.String("vpc-1")}, vpcAssociations["rslvr-rr-1"])
	a.Equal([]*string{ptr.String("vpc-2")}, vpcAssociations["rslvr-rr-2"])

	// Test case: Multiple VPC associations for a single resolver rule
	mockRoute53Resolver.EXPECT().
		ListResolverRuleAssociations(&route53resolver.ListResolverRuleAssociationsInput{}).
		Return(&route53resolver.ListResolverRuleAssociationsOutput{
			ResolverRuleAssociations: []*route53resolver.ResolverRuleAssociation{
				{
					ResolverRuleId: ptr.String("rslvr-rr-3"),
					VPCId:          ptr.String("vpc-3"),
				},
				{
					ResolverRuleId: ptr.String("rslvr-rr-3"),
					VPCId:          ptr.String("vpc-4"),
				},
			},
		}, nil)

	vpcAssociations, err = resolverRulesToVpcIDs(mockRoute53Resolver)
	a.Nil(err)
	a.NotNil(vpcAssociations)
	a.Len(vpcAssociations, 1)
	a.Equal([]*string{ptr.String("vpc-3"), ptr.String("vpc-4")}, vpcAssociations["rslvr-rr-3"])
}
