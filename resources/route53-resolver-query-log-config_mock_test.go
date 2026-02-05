package resources

import (
	"context"
	"testing"

	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	r53r "github.com/aws/aws-sdk-go-v2/service/route53resolver"
	r53rtypes "github.com/aws/aws-sdk-go-v2/service/route53resolver/types"

	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/v3/mocks/mock_route53resolverv2"
)

func Test_Mock_Route53ResolverQueryLogConfig_List(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRoute53Resolver := mock_route53resolverv2.NewMockRoute53ResolverAPI(ctrl)

	mockRoute53Resolver.EXPECT().ListResolverQueryLogConfigAssociations(gomock.Any(),
		gomock.Any()).Return(&r53r.ListResolverQueryLogConfigAssociationsOutput{
		ResolverQueryLogConfigAssociations: []r53rtypes.ResolverQueryLogConfigAssociation{
			{
				Id:                       ptr.String("rqlca-12345"),
				ResolverQueryLogConfigId: ptr.String("rqlc-12345"),
				ResourceId:               ptr.String("vpc-11111"),
				Status:                   "ACTIVE",
				CreationTime:             ptr.String("2023-01-01T00:00:00Z"),
			},
			{
				Id:                       ptr.String("rqlca-23456"),
				ResolverQueryLogConfigId: ptr.String("rqlc-67890"),
				ResourceId:               ptr.String("vpc-22222"),
				Status:                   "ACTIVE",
				CreationTime:             ptr.String("2023-01-01T01:00:00Z"),
			},
		},
	}, nil)

	mockRoute53Resolver.EXPECT().ListResolverQueryLogConfigs(gomock.Any(), gomock.Any()).Return(&r53r.ListResolverQueryLogConfigsOutput{
		ResolverQueryLogConfigs: []r53rtypes.ResolverQueryLogConfig{
			{
				Id:               ptr.String("rqlc-12345"),
				Name:             ptr.String("QueryLogConfig1"),
				Arn:              ptr.String("arn:aws:route53resolver:us-east-1:123456789012:resolver-query-log-config/rqlc-12345"),
				AssociationCount: 1,
				CreatorRequestId: ptr.String("RequestId1"),
				OwnerId:          ptr.String("123456789012"),
				ShareStatus:      r53rtypes.ShareStatusNotShared,
				DestinationArn:   ptr.String("arn:aws:s3:::query-log-bucket-1"),
				CreationTime:     ptr.String("2023-01-01T00:00:00Z"),
			},
			{
				Id:               ptr.String("rqlc-67890"),
				Name:             ptr.String("QueryLogConfig2"),
				Arn:              ptr.String("arn:aws:route53resolver:us-east-1:123456789012:resolver-query-log-config/rqlc-67890"),
				AssociationCount: 1,
				CreatorRequestId: ptr.String("RequestId2"),
				OwnerId:          ptr.String("123456789012"),
				ShareStatus:      r53rtypes.ShareStatusSharedByMe,
				DestinationArn:   ptr.String("arn:aws:s3:::query-log-bucket-2"),
				CreationTime:     ptr.String("2023-01-02T00:00:00Z"),
			},
		},
	}, nil)

	lister := &Route53ResolverQueryLogConfigLister{
		svc: mockRoute53Resolver,
	}

	resources, err := lister.List(context.TODO(), testListerOpts)
	a.Nil(err)
	a.Len(resources, 2)

	expectedResources := []resource.Resource{
		&Route53ResolverQueryLogConfig{
			svc:                    mockRoute53Resolver,
			resourceAssociationIds: []*string{ptr.String("vpc-11111")},
			AssociationCount:       1,
			Id:                     ptr.String("rqlc-12345"),
			Name:                   ptr.String("QueryLogConfig1"),
			Arn:                    ptr.String("arn:aws:route53resolver:us-east-1:123456789012:resolver-query-log-config/rqlc-12345"),
			CreatorRequestId:       ptr.String("RequestId1"),
			OwnerId:                ptr.String("123456789012"),
			ShareStatus:            r53rtypes.ShareStatusNotShared,
			DestinationArn:         ptr.String("arn:aws:s3:::query-log-bucket-1"),
			CreationTime:           ptr.String("2023-01-01T00:00:00Z"),
		},
		&Route53ResolverQueryLogConfig{
			svc:                    mockRoute53Resolver,
			resourceAssociationIds: []*string{ptr.String("vpc-22222")},
			AssociationCount:       1,
			Id:                     ptr.String("rqlc-67890"),
			Name:                   ptr.String("QueryLogConfig2"),
			Arn:                    ptr.String("arn:aws:route53resolver:us-east-1:123456789012:resolver-query-log-config/rqlc-67890"),
			CreatorRequestId:       ptr.String("RequestId2"),
			OwnerId:                ptr.String("123456789012"),
			ShareStatus:            r53rtypes.ShareStatusSharedByMe,
			DestinationArn:         ptr.String("arn:aws:s3:::query-log-bucket-2"),
			CreationTime:           ptr.String("2023-01-02T00:00:00Z"),
		},
	}

	a.Equal(expectedResources[0], resources[0])
	a.Equal(expectedResources[1], resources[1])

	a.Equal(expectedResources, resources)
}

func Test_Mock_Route53ResolverQueryLogConfig_Remove(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRoute53Resolver := mock_route53resolverv2.NewMockRoute53ResolverAPI(ctrl)

	mockRoute53Resolver.EXPECT().
		DisassociateResolverQueryLogConfig(gomock.Any(), gomock.Any()).
		Return(&r53r.DisassociateResolverQueryLogConfigOutput{}, nil)

	mockRoute53Resolver.EXPECT().
		DeleteResolverQueryLogConfig(gomock.Any(), gomock.Any()).
		Return(&r53r.DeleteResolverQueryLogConfigOutput{}, nil)

	frg := &Route53ResolverQueryLogConfig{
		svc:                    mockRoute53Resolver,
		resourceAssociationIds: []*string{ptr.String("rslvr-frgassoc-1")},
		Arn:                    ptr.String("arn:aws:route53resolver:us-east-1:123456123456:firewall-rule-group/rslvr-frg-3"),
		CreatorRequestId:       ptr.String("SomeAwsServiceCommand"),
		Id:                     ptr.String("rslvr-frg-3"),
		OwnerId:                ptr.String("Route 53 Resolver DNS Firewall"),
		Name:                   ptr.String("Internet Resolver"),
	}

	err := frg.Remove(context.TODO())
	a.Nil(err)
}

func Test_Mock_Route53ResolverQueryLogConfig_Properties(t *testing.T) {
	a := assert.New(t)

	frg := &Route53ResolverQueryLogConfig{
		Arn:              ptr.String("arn:aws:route53resolver:us-east-1:123456123456:firewall-rule-group/rslvr-frg-3"),
		CreatorRequestId: ptr.String("SomeAwsServiceCommand"),
		Id:               ptr.String("rslvr-frg-3"),
		OwnerId:          ptr.String("Route 53 Resolver DNS Firewall"),
		Name:             ptr.String("Internet Resolver"),
	}

	properties := frg.Properties()
	a.Equal("arn:aws:route53resolver:us-east-1:123456123456:firewall-rule-group/rslvr-frg-3", properties.Get("Arn"))
	a.Equal("SomeAwsServiceCommand", properties.Get("CreatorRequestId"))
	a.Equal("rslvr-frg-3", properties.Get("Id"))
	a.Equal("Route 53 Resolver DNS Firewall", properties.Get("OwnerId"))
	a.Equal("Internet Resolver", properties.Get("Name"))
}
