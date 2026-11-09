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

func Test_Mock_AppConfigExtensionAssociation_List(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mock_appconfig.NewMockAppConfigAPI(ctrl)

	extensionArn := "arn:aws:appconfig:us-east-1::extension/AWS.AppConfig.DeploymentNotificationsToEventBridge/1"
	applicationArn := "arn:aws:appconfig:us-east-1:123456789012:application/app1234"

	mockSvc.EXPECT().
		ListExtensionAssociations(gomock.Any(), &appconfig.ListExtensionAssociationsInput{
			MaxResults: ptr.Int32(50),
		}, gomock.Any()).
		Return(&appconfig.ListExtensionAssociationsOutput{
			Items: []appconfigtypes.ExtensionAssociationSummary{
				{
					Id:           ptr.String("assoc12"),
					ExtensionArn: ptr.String(extensionArn),
					ResourceArn:  ptr.String(applicationArn),
				},
			},
		}, nil)

	lister := &AppConfigExtensionAssociationLister{svc: mockSvc}

	resources, err := lister.List(context.TODO(), testListerOpts)
	a.Nil(err)
	a.Len(resources, 1)

	r := resources[0].(*AppConfigExtensionAssociation)
	a.Equal("assoc12", r.String())

	props := r.Properties()
	a.Equal("assoc12", props.Get("ID"))
	a.Equal(extensionArn, props.Get("ExtensionARN"))
	a.Equal(applicationArn, props.Get("ResourceARN"))
}

func Test_Mock_AppConfigExtensionAssociation_Remove(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mock_appconfig.NewMockAppConfigAPI(ctrl)

	mockSvc.EXPECT().
		DeleteExtensionAssociation(gomock.Any(), &appconfig.DeleteExtensionAssociationInput{
			ExtensionAssociationId: ptr.String("assoc12"),
		}).
		Return(&appconfig.DeleteExtensionAssociationOutput{}, nil)

	r := &AppConfigExtensionAssociation{
		svc: mockSvc,
		ID:  ptr.String("assoc12"),
	}

	a.Nil(r.Remove(context.TODO()))
}
