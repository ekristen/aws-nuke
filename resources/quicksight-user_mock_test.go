package resources

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/aws/aws-sdk-go/aws"                //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/quicksight" //nolint:staticcheck

	"github.com/ekristen/aws-nuke/v3/mocks/mock_quicksightiface"
)

func Test_Mock_QuicksightUser_List(t *testing.T) {
	assertions := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	accountID := testListerOpts.AccountID
	mockQuickSightAPI := mock_quicksightiface.NewMockQuickSightAPI(ctrl)

	mockQuickSightAPI.EXPECT().ListNamespaces(&quicksight.ListNamespacesInput{
		AwsAccountId: accountID,
	}).Return(&quicksight.ListNamespacesOutput{
		Namespaces: []*quicksight.NamespaceInfoV2{
			{
				Name:           aws.String("default"),
				CapacityRegion: aws.String(testListerOpts.Region.Name),
			},
		},
	}, nil)

	mockQuickSightAPI.EXPECT().ListUsersPages(&quicksight.ListUsersInput{
		AwsAccountId: accountID,
		Namespace:    aws.String("default"),
	}, gomock.Any()).DoAndReturn(
		func(_ *quicksight.ListUsersInput, fn func(*quicksight.ListUsersOutput, bool) bool) error {
			fn(&quicksight.ListUsersOutput{
				UserList: []*quicksight.User{
					{
						PrincipalId: aws.String("principal-1"),
						UserName:    aws.String("user-1"),
						Active:      aws.Bool(true),
						Role:        aws.String("ADMIN"),
					},
				},
			}, true)
			return nil
		})

	lister := QuickSightUserLister{
		quicksightService: mockQuickSightAPI,
	}

	resources, err := lister.List(context.TODO(), testListerOpts)
	assertions.Nil(err)
	assertions.Len(resources, 1)

	user := resources[0].(*QuickSightUser)
	assertions.Equal("principal-1", *user.PrincipalID)
	assertions.Equal("user-1", *user.UserName)
}

func Test_Mock_QuicksightUser_List_OutsideIdentityRegion(t *testing.T) {
	assertions := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	accountID := testListerOpts.AccountID
	mockQuickSightAPI := mock_quicksightiface.NewMockQuickSightAPI(ctrl)

	// QuickSight rejects the call and names the identity region, so there are no users to
	// list here and ListUsersPages is never reached.
	mockQuickSightAPI.EXPECT().ListNamespaces(&quicksight.ListNamespacesInput{
		AwsAccountId: accountID,
	}).Return(nil, &quicksight.AccessDeniedException{
		Message_: aws.String("Operation is being called from endpoint us-east-2, " +
			"but your identity region is us-east-1. Please use the us-east-1 endpoint."),
	})

	lister := QuickSightUserLister{
		quicksightService: mockQuickSightAPI,
	}

	resources, err := lister.List(context.TODO(), testListerOpts)
	assertions.Nil(err)
	assertions.Equal(0, len(resources))
}

func Test_Mock_QuicksightUser_Remove(t *testing.T) {
	assertions := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	accountID := testListerOpts.AccountID
	mockQuickSightAPI := mock_quicksightiface.NewMockQuickSightAPI(ctrl)

	mockQuickSightAPI.EXPECT().DeleteUserByPrincipalId(&quicksight.DeleteUserByPrincipalIdInput{
		AwsAccountId: accountID,
		Namespace:    aws.String("default"),
		PrincipalId:  aws.String("principal-1"),
	}).Return(&quicksight.DeleteUserByPrincipalIdOutput{}, nil)

	user := QuickSightUser{
		svc:         mockQuickSightAPI,
		accountID:   accountID,
		PrincipalID: aws.String("principal-1"),
		UserName:    aws.String("user-1"),
		Namespace:   aws.String("default"),
	}

	assertions.Nil(user.Remove(context.TODO()))
}
