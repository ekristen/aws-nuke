package resources

import (
	"context"
	"fmt"
	"testing"

	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	r53r "github.com/aws/aws-sdk-go-v2/service/route53resolver"
	r53rtypes "github.com/aws/aws-sdk-go-v2/service/route53resolver/types"

	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/v3/mocks/mock_route53resolverv2"
)

func Test_Mock_Route53ResolverFirewallDomainList_List(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRoute53Resolver := mock_route53resolverv2.NewMockRoute53ResolverAPI(ctrl)

	mockRoute53Resolver.EXPECT().ListFirewallDomainLists(gomock.Any(), gomock.Any()).Return(&r53r.ListFirewallDomainListsOutput{
		FirewallDomainLists: []r53rtypes.FirewallDomainListMetadata{
			{
				Arn:              ptr.String("arn:aws:route53resolver:us-east-1:123456123456:firewall-domain-list/rslvr-fdl-1"),
				CreatorRequestId: ptr.String("SomeAwsServiceCommand"),
				Id:               ptr.String("rslvr-fdl-1"),
				ManagedOwnerName: ptr.String("Route 53 Resolver DNS Firewall"),
				Name:             ptr.String("fdlNum1"),
			},
			{
				Arn:              ptr.String("arn:aws:route53resolver:us-east-1:123456123456:firewall-domain-list/rslvr-fdl-2"),
				CreatorRequestId: ptr.String("SomeAwsServiceCommand"),
				Id:               ptr.String("rslvr-fdl-2"),
				ManagedOwnerName: ptr.String("Route 53 Resolver DNS Firewall"),
				Name:             ptr.String("fdlNum2"),
			},
			{
				Arn:              ptr.String("arn:aws:route53resolver:us-east-1:123456123456:firewall-domain-list/rslvr-fdl-3"),
				CreatorRequestId: ptr.String("AWSConsole.88.1762108398672"),
				Id:               ptr.String("rslvr-fdl-3"),
				Name:             ptr.String("UserCreatedDomainList"),
			},
		},
	}, nil)

	lister := &Route53ResolverFirewallDomainListLister{
		svc: mockRoute53Resolver,
	}

	resources, err := lister.List(context.TODO(), testListerOpts)
	a.Nil(err)
	a.Len(resources, 3)

	expectedResources := []resource.Resource{
		&Route53ResolverFirewallDomainList{
			svc:              mockRoute53Resolver,
			Arn:              ptr.String("arn:aws:route53resolver:us-east-1:123456123456:firewall-domain-list/rslvr-fdl-1"),
			CreatorRequestId: ptr.String("SomeAwsServiceCommand"),
			Id:               ptr.String("rslvr-fdl-1"),
			ManagedOwnerName: ptr.String("Route 53 Resolver DNS Firewall"),
			Name:             ptr.String("fdlNum1"),
		},
		&Route53ResolverFirewallDomainList{
			svc:              mockRoute53Resolver,
			Arn:              ptr.String("arn:aws:route53resolver:us-east-1:123456123456:firewall-domain-list/rslvr-fdl-2"),
			CreatorRequestId: ptr.String("SomeAwsServiceCommand"),
			Id:               ptr.String("rslvr-fdl-2"),
			ManagedOwnerName: ptr.String("Route 53 Resolver DNS Firewall"),
			Name:             ptr.String("fdlNum2"),
		},
		&Route53ResolverFirewallDomainList{
			svc:              mockRoute53Resolver,
			Arn:              ptr.String("arn:aws:route53resolver:us-east-1:123456123456:firewall-domain-list/rslvr-fdl-3"),
			CreatorRequestId: ptr.String("AWSConsole.88.1762108398672"),
			Id:               ptr.String("rslvr-fdl-3"),
			Name:             ptr.String("UserCreatedDomainList"),
		},
	}

	a.Equal(expectedResources, resources)
}

func Test_Mock_Route53ResolverFirewallDomainList_Filter(t *testing.T) {
	a := assert.New(t)

	cases := []struct {
		Arn              string
		CreatorRequestId string
		Id               string
		ManagedOwnerName string
		Name             string
		Filtered         bool
	}{
		{
			Arn:              "arn:aws:route53resolver:us-east-1:123456123456:firewall-domain-list/rslvr-fdl-2",
			CreatorRequestId: "SomeAwsServiceCommand",
			Id:               "rslvr-fdl-1",
			ManagedOwnerName: "Route 53 Resolver DNS Firewall",
			Name:             "fdlNum2",
			Filtered:         true,
		},
		{
			Arn:              "arn:aws:route53resolver:us-east-1:123456123456:firewall-domain-list/rslvr-fdl-3",
			CreatorRequestId: "SomeAwsServiceCommand",
			Id:               "rslvr-fdl-2",
			Name:             "UserCreatedDomainList",
			Filtered:         false,
		},
	}

	for _, c := range cases {
		name := c.Name
		if c.Filtered {
			name = fmt.Sprintf("filtered/%s", name)
		} else {
			name = fmt.Sprintf("not-filtered/%s", name)
		}

		t.Run(name, func(t *testing.T) {
			fdl := &Route53ResolverFirewallDomainList{
				Arn:              ptr.String(c.Arn),
				CreatorRequestId: ptr.String(c.CreatorRequestId),
				Id:               ptr.String(c.Id),
				ManagedOwnerName: ptr.String(c.ManagedOwnerName),
				Name:             ptr.String(c.Name),
			}

			err := fdl.Filter()
			if c.Filtered {
				a.NotNil(err)
			} else {
				a.Nil(err)
			}
		})
	}
}

func Test_Mock_Route53ResolverFirewallDomainList_Remove(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRoute53Resolver := mock_route53resolverv2.NewMockRoute53ResolverAPI(ctrl)

	mockRoute53Resolver.EXPECT().
		DeleteFirewallDomainList(gomock.Any(), gomock.Any()).
		Return(&r53r.DeleteFirewallDomainListOutput{}, nil)

	fdl := &Route53ResolverFirewallDomainList{
		svc:              mockRoute53Resolver,
		Arn:              ptr.String("arn:aws:route53resolver:us-east-1:123456123456:firewall-domain-list/rslvr-fdl-3"),
		CreatorRequestId: ptr.String("SomeAwsServiceCommand"),
		Id:               ptr.String("rslvr-fdl-3"),
		ManagedOwnerName: ptr.String("Route 53 Resolver DNS Firewall"),
		Name:             ptr.String("Internet Resolver"),
	}

	err := fdl.Remove(context.TODO())
	a.Nil(err)
}

func Test_Mock_Route53ResolverFirewallDomainList_Properties(t *testing.T) {
	a := assert.New(t)

	fdl := &Route53ResolverFirewallDomainList{
		Arn:              ptr.String("arn:aws:route53resolver:us-east-1:123456123456:firewall-domain-list/rslvr-fdl-3"),
		CreatorRequestId: ptr.String("SomeAwsServiceCommand"),
		Id:               ptr.String("rslvr-fdl-3"),
		ManagedOwnerName: ptr.String("Route 53 Resolver DNS Firewall"),
		Name:             ptr.String("Internet Resolver"),
	}

	properties := fdl.Properties()
	a.Equal("arn:aws:route53resolver:us-east-1:123456123456:firewall-domain-list/rslvr-fdl-3", properties.Get("Arn"))
	a.Equal("SomeAwsServiceCommand", properties.Get("CreatorRequestId"))
	a.Equal("rslvr-fdl-3", properties.Get("Id"))
	a.Equal("Route 53 Resolver DNS Firewall", properties.Get("ManagedOwnerName"))
	a.Equal("Internet Resolver", properties.Get("Name"))
}
