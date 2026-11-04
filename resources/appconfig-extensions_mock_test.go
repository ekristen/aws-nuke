package resources

import (
	"context"
	"testing"

	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/aws/aws-sdk-go-v2/service/appconfig"
	appconfigtypes "github.com/aws/aws-sdk-go-v2/service/appconfig/types"

	"github.com/ekristen/aws-nuke/v3/mocks/mock_appconfig"
)

func Test_Mock_AppConfigExtension_List(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mock_appconfig.NewMockAppConfigAPI(ctrl)

	gomock.InOrder(
		mockSvc.EXPECT().
			ListExtensions(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(&appconfig.ListExtensionsOutput{
				Items: []appconfigtypes.ExtensionSummary{
					{
						Id:            ptr.String("abc1234"),
						Name:          ptr.String("my-extension"),
						Arn:           ptr.String("arn:aws:appconfig:us-east-1:123456789012:extension/abc1234/2"),
						VersionNumber: 2,
					},
				},
				NextToken: ptr.String("next"),
			}, nil),
		mockSvc.EXPECT().
			ListExtensions(gomock.Any(), &appconfig.ListExtensionsInput{
				MaxResults: ptr.Int32(50),
				NextToken:  ptr.String("next"),
			}, gomock.Any()).
			Return(&appconfig.ListExtensionsOutput{
				Items: []appconfigtypes.ExtensionSummary{
					{
						Id:            ptr.String("AWS.AppConfig.FeatureFlags"),
						Name:          ptr.String("AWS.AppConfig.FeatureFlags"),
						Arn:           ptr.String("arn:aws:appconfig:us-east-1::extension/AWS.AppConfig.FeatureFlags/1"),
						VersionNumber: 1,
					},
				},
			}, nil),
	)

	lister := &AppConfigExtensionLister{svc: mockSvc}

	resources, err := lister.List(context.TODO(), testListerOpts)
	a.Nil(err)
	a.Len(resources, 2)

	custom := resources[0].(*AppConfigExtension)
	a.Equal("my-extension", custom.String())
	a.Equal("abc1234", *custom.ID)
	a.Equal(int32(2), *custom.VersionNumber)
	a.Nil(custom.Filter())

	awsAuthored := resources[1].(*AppConfigExtension)
	a.Equal("AWS.AppConfig.FeatureFlags", awsAuthored.String())
	a.EqualError(awsAuthored.Filter(), "cannot delete AWS authored extension")
}

func Test_Mock_AppConfigExtension_Remove(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mock_appconfig.NewMockAppConfigAPI(ctrl)

	mockSvc.EXPECT().
		DeleteExtension(gomock.Any(), &appconfig.DeleteExtensionInput{
			ExtensionIdentifier: ptr.String("abc1234"),
			VersionNumber:       ptr.Int32(2),
		}).
		Return(&appconfig.DeleteExtensionOutput{}, nil)

	r := &AppConfigExtension{
		svc:           mockSvc,
		ID:            ptr.String("abc1234"),
		Name:          ptr.String("my-extension"),
		VersionNumber: ptr.Int32(2),
	}

	a.Nil(r.Remove(context.TODO()))
}

func Test_AppConfigExtension_Properties(t *testing.T) {
	a := assert.New(t)

	r := &AppConfigExtension{
		ID:            ptr.String("abc1234"),
		Name:          ptr.String("my-extension"),
		ARN:           ptr.String("arn:aws:appconfig:us-east-1:123456789012:extension/abc1234/2"),
		VersionNumber: ptr.Int32(2),
	}

	props := r.Properties()
	a.Equal("abc1234", props.Get("ID"))
	a.Equal("my-extension", props.Get("Name"))
	a.Equal("arn:aws:appconfig:us-east-1:123456789012:extension/abc1234/2", props.Get("ARN"))
	a.Equal("2", props.Get("VersionNumber"))
}
