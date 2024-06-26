//go:generate ../mocks/generate_mocks.sh quicksight quicksightiface
//go:generate ../mocks/generate_mocks.sh sts stsiface

package resources

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/quicksight"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/ekristen/aws-nuke/v3/mocks/mock_quicksightiface"
	"github.com/ekristen/aws-nuke/v3/mocks/mock_stsiface"
	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
	libsettings "github.com/ekristen/libnuke/pkg/settings"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func Test_Mock_QuicksightSubscription_List_ValidSubscription(t *testing.T) {
	assertions := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	accountID := "123456789012"
	quickSightAccountInfo := quicksight.AccountInfo{
		AccountName:               aws.String("AccountName"),
		NotificationEmail:         aws.String("notification@email.com"),
		Edition:                   aws.String("Edition"),
		AccountSubscriptionStatus: aws.String("ACCOUNT_CREATED"),
	}
	mockQuickSightAPI := mock_quicksightiface.NewMockQuickSightAPI(ctrl)
	mockSTSAPI := mock_stsiface.NewMockSTSAPI(ctrl)

	mockSTSAPI.EXPECT().GetCallerIdentity(&sts.GetCallerIdentityInput{}).Return(&sts.GetCallerIdentityOutput{
		Account: &accountID,
	}, nil)

	mockQuickSightAPI.EXPECT().DescribeAccountSubscription(&quicksight.DescribeAccountSubscriptionInput{
		AwsAccountId: &accountID,
	}).Return(&quicksight.DescribeAccountSubscriptionOutput{
		AccountInfo: &quickSightAccountInfo,
	}, nil)

	quicksightSubscriptionListener := QuickSightSubscriptionLister{
		quicksightService: mockQuickSightAPI,
		stsService:        mockSTSAPI,
	}

	resources, err := quicksightSubscriptionListener.List(context.TODO(), &nuke.ListerOpts{})
	assertions.Nil(err)

	resource := resources[0].(*QuickSightSubscription)
	assertions.Equal(*resource.accountID, accountID)
	assertions.Equal(resource.edition, quickSightAccountInfo.Edition)
	assertions.Equal(resource.name, quickSightAccountInfo.AccountName)
	assertions.Equal(resource.notificationEmail, quickSightAccountInfo.NotificationEmail)
	assertions.Equal(resource.status, quickSightAccountInfo.AccountSubscriptionStatus)
}

func Test_Mock_QuicksightSubscription_List_SubscriptionNotFound(t *testing.T) {
	assertions := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	accountID := "123456789012"
	quickSightSubscriptionNotFoundError := &quicksight.ResourceNotFoundException{
		Message_: aws.String("Resource not found"),
	}

	mockQuickSightAPI := mock_quicksightiface.NewMockQuickSightAPI(ctrl)
	mockSTSAPI := mock_stsiface.NewMockSTSAPI(ctrl)

	mockSTSAPI.EXPECT().GetCallerIdentity(&sts.GetCallerIdentityInput{}).Return(&sts.GetCallerIdentityOutput{
		Account: &accountID,
	}, nil)

	mockQuickSightAPI.EXPECT().DescribeAccountSubscription(&quicksight.DescribeAccountSubscriptionInput{
		AwsAccountId: &accountID,
	}).Return(nil, quickSightSubscriptionNotFoundError)

	quicksightSubscriptionListener := QuickSightSubscriptionLister{
		quicksightService: mockQuickSightAPI,
		stsService:        mockSTSAPI,
	}

	resources, err := quicksightSubscriptionListener.List(context.TODO(), &nuke.ListerOpts{})
	assertions.Nil(err)
	assertions.Equal(0, len(resources))
}

func Test_Mock_QuicksightSubscription_List_ErrorOnSTS(t *testing.T) {
	assertions := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockQuickSightAPI := mock_quicksightiface.NewMockQuickSightAPI(ctrl)
	mockSTSAPI := mock_stsiface.NewMockSTSAPI(ctrl)

	mockSTSAPI.EXPECT().GetCallerIdentity(&sts.GetCallerIdentityInput{}).Return(nil, errors.New("MOCK_ERROR"))

	quicksightSubscriptionListener := QuickSightSubscriptionLister{
		quicksightService: mockQuickSightAPI,
		stsService:        mockSTSAPI,
	}

	resources, err := quicksightSubscriptionListener.List(context.TODO(), &nuke.ListerOpts{})
	assertions.EqualError(err, "MOCK_ERROR")
	assertions.Nil(resources)
}

func Test_Mock_QuicksightSubscription_List_ErrorOnQuicksight(t *testing.T) {
	assertions := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	accountID := "123456789012"

	mockQuickSightAPI := mock_quicksightiface.NewMockQuickSightAPI(ctrl)
	mockSTSAPI := mock_stsiface.NewMockSTSAPI(ctrl)

	mockSTSAPI.EXPECT().GetCallerIdentity(&sts.GetCallerIdentityInput{}).Return(&sts.GetCallerIdentityOutput{
		Account: &accountID,
	}, nil)

	mockQuickSightAPI.EXPECT().DescribeAccountSubscription(&quicksight.DescribeAccountSubscriptionInput{
		AwsAccountId: &accountID,
	}).Return(nil, errors.New("MOCK_ERROR"))

	quicksightSubscriptionListener := QuickSightSubscriptionLister{
		quicksightService: mockQuickSightAPI,
		stsService:        mockSTSAPI,
	}

	resources, err := quicksightSubscriptionListener.List(context.TODO(), &nuke.ListerOpts{})
	assertions.EqualError(err, "MOCK_ERROR")
	assertions.Nil(resources)
}

func Test_Mock_QuicksightSubscription_Remove_No_Settings(t *testing.T) {
	assertions := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	accountID := "123456789012"
	subscriptionName := aws.String("Name")
	subscriptionDefaultNamespace := aws.String("Default")
	subscriptionNotificationEmail := aws.String("notification@email.com")
	subscriptionEdition := aws.String("Edition")
	subscriptionStatus := aws.String("ACCOUNT_CREATED")

	mockQuickSightAPI := mock_quicksightiface.NewMockQuickSightAPI(ctrl)

	mockQuickSightAPI.EXPECT().DescribeAccountSettings(&quicksight.DescribeAccountSettingsInput{
		AwsAccountId: &accountID,
	}).Return(&quicksight.DescribeAccountSettingsOutput{
		AccountSettings: &quicksight.AccountSettings{
			DefaultNamespace:             subscriptionDefaultNamespace,
			NotificationEmail:            subscriptionNotificationEmail,
			TerminationProtectionEnabled: aws.Bool(true),
		},
	}, nil).Times(0)

	mockQuickSightAPI.EXPECT().UpdateAccountSettings(&quicksight.UpdateAccountSettingsInput{
		AwsAccountId:                 &accountID,
		DefaultNamespace:             subscriptionDefaultNamespace,
		NotificationEmail:            subscriptionNotificationEmail,
		TerminationProtectionEnabled: aws.Bool(false),
	}).Return(&quicksight.UpdateAccountSettingsOutput{}, nil).Times(0)

	mockQuickSightAPI.EXPECT().DeleteAccountSubscription(&quicksight.DeleteAccountSubscriptionInput{
		AwsAccountId: &accountID,
	}).Return(&quicksight.DeleteAccountSubscriptionOutput{}, nil).Times(1)

	quicksightSubscription := QuickSightSubscription{
		svc:               mockQuickSightAPI,
		accountID:         &accountID,
		name:              subscriptionName,
		notificationEmail: subscriptionNotificationEmail,
		edition:           subscriptionEdition,
		status:            subscriptionStatus,
	}

	err := quicksightSubscription.Remove(context.TODO())

	assertions.Nil(err)
}

func Test_Mock_QuicksightSubscription_Remove_TerminationSetting_Is_False(t *testing.T) {
	assertions := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	accountID := "123456789012"
	subscriptionName := aws.String("Name")
	subscriptionDefaultNamespace := aws.String("Default")
	subscriptionNotificationEmail := aws.String("notification@email.com")
	subscriptionEdition := aws.String("Edition")
	subscriptionStatus := aws.String("ACCOUNT_CREATED")
	settings := libsettings.Setting{}
	settings.Set("DisableTerminationProtection", false)

	mockQuickSightAPI := mock_quicksightiface.NewMockQuickSightAPI(ctrl)

	mockQuickSightAPI.EXPECT().DescribeAccountSettings(&quicksight.DescribeAccountSettingsInput{
		AwsAccountId: &accountID,
	}).Return(&quicksight.DescribeAccountSettingsOutput{
		AccountSettings: &quicksight.AccountSettings{
			DefaultNamespace:             subscriptionDefaultNamespace,
			NotificationEmail:            subscriptionNotificationEmail,
			TerminationProtectionEnabled: aws.Bool(true),
		},
	}, nil).Times(0)

	mockQuickSightAPI.EXPECT().UpdateAccountSettings(&quicksight.UpdateAccountSettingsInput{
		AwsAccountId:                 &accountID,
		DefaultNamespace:             subscriptionDefaultNamespace,
		NotificationEmail:            subscriptionNotificationEmail,
		TerminationProtectionEnabled: aws.Bool(false),
	}).Return(&quicksight.UpdateAccountSettingsOutput{}, nil).Times(0)

	mockQuickSightAPI.EXPECT().DeleteAccountSubscription(&quicksight.DeleteAccountSubscriptionInput{
		AwsAccountId: &accountID,
	}).Return(&quicksight.DeleteAccountSubscriptionOutput{}, nil).Times(1)

	quicksightSubscription := QuickSightSubscription{
		svc:               mockQuickSightAPI,
		accountID:         &accountID,
		name:              subscriptionName,
		notificationEmail: subscriptionNotificationEmail,
		edition:           subscriptionEdition,
		status:            subscriptionStatus,
		settings:          &settings,
	}

	err := quicksightSubscription.Remove(context.TODO())

	assertions.Nil(err)
}

func Test_Mock_QuicksightSubscription_Remove_TerminationSetting_Is_True(t *testing.T) {
	assertions := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	accountID := "123456789012"
	subscriptionName := aws.String("Name")
	subscriptionDefaultNamespace := aws.String("Default")
	subscriptionNotificationEmail := aws.String("notification@email.com")
	subscriptionEdition := aws.String("Edition")
	subscriptionStatus := aws.String("ACCOUNT_CREATED")
	settings := libsettings.Setting{}
	settings.Set("DisableTerminationProtection", true)

	mockQuickSightAPI := mock_quicksightiface.NewMockQuickSightAPI(ctrl)

	mockQuickSightAPI.EXPECT().DescribeAccountSettings(&quicksight.DescribeAccountSettingsInput{
		AwsAccountId: &accountID,
	}).Return(&quicksight.DescribeAccountSettingsOutput{
		AccountSettings: &quicksight.AccountSettings{
			DefaultNamespace:             subscriptionDefaultNamespace,
			NotificationEmail:            subscriptionNotificationEmail,
			TerminationProtectionEnabled: aws.Bool(true),
		},
	}, nil).Times(1)

	mockQuickSightAPI.EXPECT().UpdateAccountSettings(&quicksight.UpdateAccountSettingsInput{
		AwsAccountId:                 &accountID,
		DefaultNamespace:             subscriptionDefaultNamespace,
		NotificationEmail:            subscriptionNotificationEmail,
		TerminationProtectionEnabled: aws.Bool(false),
	}).Return(&quicksight.UpdateAccountSettingsOutput{}, nil).Times(1)

	mockQuickSightAPI.EXPECT().DeleteAccountSubscription(&quicksight.DeleteAccountSubscriptionInput{
		AwsAccountId: &accountID,
	}).Return(&quicksight.DeleteAccountSubscriptionOutput{}, nil).Times(1)

	quicksightSubscription := QuickSightSubscription{
		svc:               mockQuickSightAPI,
		accountID:         &accountID,
		name:              subscriptionName,
		notificationEmail: subscriptionNotificationEmail,
		edition:           subscriptionEdition,
		status:            subscriptionStatus,
		settings:          &settings,
	}

	err := quicksightSubscription.Remove(context.TODO())

	assertions.Nil(err)
}

func Test_Mock_QuicksightSubscription_Filter(t *testing.T) {
	assertions := assert.New(t)

	accountID := "123456789012"
	subscriptionName := aws.String("Name")
	subscriptionNotificationEmail := aws.String("notification@email.com")
	subscriptionEdition := aws.String("Edition")
	subscriptionStatus := aws.String("ACCOUNT_CREATED")

	quicksightSubscription := QuickSightSubscription{
		accountID:         &accountID,
		name:              subscriptionName,
		notificationEmail: subscriptionNotificationEmail,
		edition:           subscriptionEdition,
		status:            subscriptionStatus,
	}

	err := quicksightSubscription.Filter()

	assertions.Nil(err)
}

func Test_Mock_QuicksightSubscription_Filter_Status(t *testing.T) {
	assertions := assert.New(t)

	accountID := "123456789012"
	subscriptionName := aws.String("Name")
	subscriptionNotificationEmail := aws.String("notification@email.com")
	subscriptionEdition := aws.String("Edition")
	subscriptionStatus := aws.String("UNSUBSCRIBED")

	quicksightSubscription := QuickSightSubscription{
		accountID:         &accountID,
		name:              subscriptionName,
		notificationEmail: subscriptionNotificationEmail,
		edition:           subscriptionEdition,
		status:            subscriptionStatus,
	}

	err := quicksightSubscription.Filter()

	assertions.EqualError(err, "subscription is not active")
}

func Test_Mock_QuicksightSubscription_Filter_Name(t *testing.T) {
	assertions := assert.New(t)

	accountID := "123456789012"
	subscriptionName := aws.String("NOT_AVAILABLE")
	subscriptionNotificationEmail := aws.String("notification@email.com")
	subscriptionEdition := aws.String("Edition")
	subscriptionStatus := aws.String("ACCOUNT_CREATED")

	quicksightSubscription := QuickSightSubscription{
		accountID:         &accountID,
		name:              subscriptionName,
		notificationEmail: subscriptionNotificationEmail,
		edition:           subscriptionEdition,
		status:            subscriptionStatus,
	}

	err := quicksightSubscription.Filter()

	assertions.EqualError(err, "subscription name is not available yet")
}
