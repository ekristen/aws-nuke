package resources

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"

	"github.com/ekristen/aws-nuke/v3/mocks/mock_iamiface"
	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

func Test_Mock_IAMUser_List(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockIAM := mock_iamiface.NewMockIAMAPI(ctrl)

	mockIAM.EXPECT().GetUser(&iam.GetUserInput{
		UserName: ptr.String("foo"),
	}).Return(&iam.GetUserOutput{
		User: &iam.User{
			UserName: ptr.String("foo"),
		},
	}, nil)

	mockIAM.EXPECT().ListUsersPages(gomock.Any(), gomock.Any()).
		Do(func(_ *iam.ListUsersInput, fn func(page *iam.ListUsersOutput, lastPage bool) bool) {
			fn(&iam.ListUsersOutput{
				Users: []*iam.User{
					{
						UserName: ptr.String("foo"),
					},
				},
			}, true)
		}).Return(nil)

	lister := IAMUserLister{
		mockSvc: mockIAM,
	}

	resources, err := lister.List(context.TODO(), &nuke.ListerOpts{
		Region: &nuke.Region{
			Name: "us-east-2",
		},
		Session: session.Must(session.NewSession()),
	})
	a.NoError(err)
	a.Len(resources, 1)
}

func Test_Mock_IAMUser_Remove(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockIAM := mock_iamiface.NewMockIAMAPI(ctrl)

	mockIAM.EXPECT().DeleteUserPermissionsBoundary(&iam.DeleteUserPermissionsBoundaryInput{
		UserName: ptr.String("foobar"),
	}).Return(&iam.DeleteUserPermissionsBoundaryOutput{}, nil)

	mockIAM.EXPECT().DeleteUser(gomock.Eq(&iam.DeleteUserInput{
		UserName: ptr.String("foobar"),
	})).Return(&iam.DeleteUserOutput{}, nil)

	iamUser := IAMUser{
		svc:                   mockIAM,
		name:                  ptr.String("foobar"),
		hasPermissionBoundary: true,
	}

	err := iamUser.Remove(context.TODO())
	a.Nil(err)
}
