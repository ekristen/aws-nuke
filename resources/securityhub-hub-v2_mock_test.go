package resources

import (
	"context"
	"testing"

	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/aws/aws-sdk-go-v2/service/securityhub"
	securityhubtypes "github.com/aws/aws-sdk-go-v2/service/securityhub/types"

	"github.com/ekristen/aws-nuke/v3/mocks/mock_securityhub"
)

const testSecurityHubV2Arn = "arn:aws:securityhub:us-east-1:123456123456:hub-v2/default"

func Test_Mock_SecurityHubV2_List(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mock_securityhub.NewMockSecurityHubV2API(ctrl)

	mockSvc.EXPECT().DescribeSecurityHubV2(gomock.Any(), gomock.Any()).
		Return(&securityhub.DescribeSecurityHubV2Output{
			HubV2Arn:     ptr.String(testSecurityHubV2Arn),
			SubscribedAt: ptr.String("2025-09-01T12:00:00.000Z"),
			Features: map[string]securityhubtypes.FeatureDetail{
				string(securityhubtypes.FeatureNameNetworkScanning): {
					FeatureStatus: securityhubtypes.FeatureStatusEnabled,
				},
			},
		}, nil)

	lister := &SecurityHubV2Lister{svc: mockSvc}

	resources, err := lister.List(context.TODO(), testListerOpts)
	a.Nil(err)
	a.Len(resources, 1)

	hub := resources[0].(*SecurityHubV2)
	a.Equal(testSecurityHubV2Arn, *hub.HubV2ARN)
	a.Equal("2025-09-01T12:00:00.000Z", *hub.SubscribedAt)
	a.Equal(map[string]string{"NETWORK_SCANNING": "ENABLED"}, hub.Features)
}

// Test_Mock_SecurityHubV2_List_NotEnabled asserts that regions where Security Hub V2 was never enabled, or is not
// available, yield no resources rather than failing the scan.
func Test_Mock_SecurityHubV2_List_NotEnabled(t *testing.T) {
	cases := []struct {
		name string
		err  error
	}{
		{name: "ResourceNotFound", err: &securityhubtypes.ResourceNotFoundException{}},
		{name: "InvalidAccess", err: &securityhubtypes.InvalidAccessException{}},
		{name: "AccessDenied", err: &securityhubtypes.AccessDeniedException{}},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			a := assert.New(t)
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSvc := mock_securityhub.NewMockSecurityHubV2API(ctrl)

			mockSvc.EXPECT().DescribeSecurityHubV2(gomock.Any(), gomock.Any()).Return(nil, c.err)

			lister := &SecurityHubV2Lister{svc: mockSvc}

			resources, err := lister.List(context.TODO(), testListerOpts)
			a.Nil(err)
			a.Len(resources, 0)
		})
	}
}

func Test_Mock_SecurityHubV2_List_Error(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mock_securityhub.NewMockSecurityHubV2API(ctrl)

	mockSvc.EXPECT().DescribeSecurityHubV2(gomock.Any(), gomock.Any()).
		Return(nil, &securityhubtypes.InternalException{})

	lister := &SecurityHubV2Lister{svc: mockSvc}

	resources, err := lister.List(context.TODO(), testListerOpts)
	a.Error(err)
	a.Nil(resources)
}

func Test_Mock_SecurityHubV2_Remove(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mock_securityhub.NewMockSecurityHubV2API(ctrl)

	mockSvc.EXPECT().DisableSecurityHubV2(gomock.Any(), gomock.Any()).
		Return(&securityhub.DisableSecurityHubV2Output{}, nil)

	hub := &SecurityHubV2{
		svc:      mockSvc,
		HubV2ARN: ptr.String(testSecurityHubV2Arn),
	}

	err := hub.Remove(context.TODO())
	a.Nil(err)
}

func Test_Mock_SecurityHubV2_Properties(t *testing.T) {
	a := assert.New(t)

	hub := &SecurityHubV2{
		HubV2ARN:     ptr.String(testSecurityHubV2Arn),
		SubscribedAt: ptr.String("2025-09-01T12:00:00.000Z"),
		Features: map[string]string{
			"NETWORK_SCANNING": "ENABLED",
		},
	}

	props := hub.Properties()
	a.Equal(testSecurityHubV2Arn, props.Get("HubV2ARN"))
	a.Equal("2025-09-01T12:00:00.000Z", props.Get("SubscribedAt"))
	a.Equal("ENABLED", props.Get("feature:NETWORK_SCANNING"))
	a.Equal(testSecurityHubV2Arn, hub.String())
}
