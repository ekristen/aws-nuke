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
	"github.com/ekristen/aws-nuke/mocks/mock_quicksightiface"
	"github.com/ekristen/aws-nuke/mocks/mock_stsiface"
	"github.com/ekristen/aws-nuke/pkg/nuke"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func Test_Mock_QuicksightSubscription_List_ValidSubscription(t *testing.T) {
	assert := assert.New(t)
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
	assert.Nil(err)

	resource := resources[0].(*QuicksightSubscription)
	assert.Equal(*resource.accountId, accountID)
	assert.Equal(resource.edition, quickSightAccountInfo.Edition)
	assert.Equal(resource.name, quickSightAccountInfo.AccountName)
	assert.Equal(resource.notificationEmail, quickSightAccountInfo.NotificationEmail)
	assert.Equal(resource.status, quickSightAccountInfo.AccountSubscriptionStatus)
}

func Test_Mock_QuicksightSubscription_List_SubscriptionNotFound(t *testing.T) {
	assert := assert.New(t)
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
	assert.Nil(err)
	assert.Equal(0, len(resources))
}

func Test_Mock_QuicksightSubscription_List_ErrorOnSTS(t *testing.T) {
	assert := assert.New(t)
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
	assert.EqualError(err, "MOCK_ERROR")
	assert.Nil(resources)
}

func Test_Mock_QuicksightSubscription_List_ErrorOnQuicksight(t *testing.T) {
	assert := assert.New(t)
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
	assert.EqualError(err, "MOCK_ERROR")
	assert.Nil(resources)
}

func Test_Mock_QuicksightSubscription_Remove(t *testing.T) {
	assert := assert.New(t)
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
	}, nil)

	mockQuickSightAPI.EXPECT().UpdateAccountSettings(&quicksight.UpdateAccountSettingsInput{
		AwsAccountId:                 &accountID,
		DefaultNamespace:             subscriptionDefaultNamespace,
		NotificationEmail:            subscriptionNotificationEmail,
		TerminationProtectionEnabled: aws.Bool(false),
	}).Return(&quicksight.UpdateAccountSettingsOutput{}, nil).Times(1)

	mockQuickSightAPI.EXPECT().DeleteAccountSubscription(&quicksight.DeleteAccountSubscriptionInput{
		AwsAccountId: &accountID,
	}).Return(&quicksight.DeleteAccountSubscriptionOutput{}, nil).Times(1)

	quicksightSubscription := QuicksightSubscription{
		svc:               mockQuickSightAPI,
		accountId:         &accountID,
		name:              subscriptionName,
		notificationEmail: subscriptionNotificationEmail,
		edition:           subscriptionEdition,
		status:            subscriptionStatus,
	}

	err := quicksightSubscription.Remove(context.TODO())

	assert.Nil(err)
}

func Test_Mock_QuicksightSubscription_NoTerminationUpdatedNeeded(t *testing.T) {
	assert := assert.New(t)
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
			TerminationProtectionEnabled: aws.Bool(false),
		},
	}, nil)

	mockQuickSightAPI.EXPECT().UpdateAccountSettings(&quicksight.UpdateAccountSettingsInput{
		AwsAccountId:                 &accountID,
		DefaultNamespace:             subscriptionDefaultNamespace,
		NotificationEmail:            subscriptionNotificationEmail,
		TerminationProtectionEnabled: aws.Bool(false),
	}).Return(&quicksight.UpdateAccountSettingsOutput{}, nil).Times(0)

	mockQuickSightAPI.EXPECT().DeleteAccountSubscription(&quicksight.DeleteAccountSubscriptionInput{
		AwsAccountId: &accountID,
	}).Return(&quicksight.DeleteAccountSubscriptionOutput{}, nil).Times(1)

	quicksightSubscription := QuicksightSubscription{
		svc:               mockQuickSightAPI,
		accountId:         &accountID,
		name:              subscriptionName,
		notificationEmail: subscriptionNotificationEmail,
		edition:           subscriptionEdition,
		status:            subscriptionStatus,
	}

	err := quicksightSubscription.Remove(context.TODO())

	assert.Nil(err)
}

func Test_Mock_QuicksightSubscription_Filter(t *testing.T) {
	assert := assert.New(t)

	accountID := "123456789012"
	subscriptionName := aws.String("Name")
	subscriptionNotificationEmail := aws.String("notification@email.com")
	subscriptionEdition := aws.String("Edition")
	subscriptionStatus := aws.String("ACCOUNT_CREATED")

	quicksightSubscription := QuicksightSubscription{
		accountId:         &accountID,
		name:              subscriptionName,
		notificationEmail: subscriptionNotificationEmail,
		edition:           subscriptionEdition,
		status:            subscriptionStatus,
	}

	err := quicksightSubscription.Filter()

	assert.Nil(err)
}

func Test_Mock_QuicksightSubscription_Filter_Status(t *testing.T) {
	assert := assert.New(t)

	accountID := "123456789012"
	subscriptionName := aws.String("Name")
	subscriptionNotificationEmail := aws.String("notification@email.com")
	subscriptionEdition := aws.String("Edition")
	subscriptionStatus := aws.String("UNSUBSCRIBED")

	quicksightSubscription := QuicksightSubscription{
		accountId:         &accountID,
		name:              subscriptionName,
		notificationEmail: subscriptionNotificationEmail,
		edition:           subscriptionEdition,
		status:            subscriptionStatus,
	}

	err := quicksightSubscription.Filter()

	assert.EqualError(err, "subscription is not active")
}

func Test_Mock_QuicksightSubscription_Filter_Name(t *testing.T) {
	assert := assert.New(t)

	accountID := "123456789012"
	subscriptionName := aws.String("NOT_AVAILABLE")
	subscriptionNotificationEmail := aws.String("notification@email.com")
	subscriptionEdition := aws.String("Edition")
	subscriptionStatus := aws.String("ACCOUNT_CREATED")

	quicksightSubscription := QuicksightSubscription{
		accountId:         &accountID,
		name:              subscriptionName,
		notificationEmail: subscriptionNotificationEmail,
		edition:           subscriptionEdition,
		status:            subscriptionStatus,
	}

	err := quicksightSubscription.Filter()

	assert.EqualError(err, "subscription name is not available yet")
}
