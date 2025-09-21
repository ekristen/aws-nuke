package resources

import (
	"context"
	"testing"
	"time"

	libsettings "github.com/ekristen/libnuke/pkg/settings"
	"github.com/golang/mock/gomock"
	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"

	"github.com/aws/aws-sdk-go/service/iam"

	"github.com/ekristen/aws-nuke/v3/mocks/mock_iamiface"
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

	resources, err := lister.List(context.TODO(), testListerOpts)
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
		Name:                  ptr.String("foobar"),
		HasPermissionBoundary: true,
		settings:              &libsettings.Setting{},
	}

	err := iamUser.Remove(context.TODO())
	a.Nil(err)
}

func Test_Mock_IAMUser_Properties(t *testing.T) {
	a := assert.New(t)

	now := time.Now().UTC()

	iamUser := IAMUser{
		Name:             ptr.String("foo"),
		Path:             ptr.String("/foo"),
		UserID:           ptr.String("foobar"),
		CreateDate:       ptr.Time(now),
		PasswordLastUsed: ptr.Time(now),
		Tags: []*iam.Tag{
			{
				Key:   ptr.String("foo"),
				Value: ptr.String("bar"),
			},
		},
		HasPermissionBoundary:  true,
		PermissionBoundaryARN:  ptr.String("arn:aws:iam::123456789012:policy/foo"),
		PermissionBoundaryType: ptr.String("PermissionsBoundary"),
	}

	a.Equal("foo", iamUser.String())
	a.Equal("foobar", iamUser.Properties().Get("UserID"))
	a.Equal("foo", iamUser.Properties().Get("Name"))
	a.Equal("true", iamUser.Properties().Get("HasPermissionBoundary"))
	a.Equal(now.Format(time.RFC3339), iamUser.Properties().Get("CreateDate"))
	a.Equal(now.Format(time.RFC3339), iamUser.Properties().Get("PasswordLastUsed"))
	a.Equal("arn:aws:iam::123456789012:policy/foo", iamUser.Properties().Get("PermissionBoundaryARN"))
	a.Equal("PermissionsBoundary", iamUser.Properties().Get("PermissionBoundaryType"))
	a.Equal("bar", iamUser.Properties().Get("tag:foo"))
	a.Equal("/foo", iamUser.Properties().Get("Path"))
}
