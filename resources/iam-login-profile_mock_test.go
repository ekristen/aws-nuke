package resources

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"

	"github.com/aws/aws-sdk-go/service/iam"

	"github.com/ekristen/aws-nuke/v3/mocks/mock_iamiface"
)

func Test_Mock_IAMLoginProfile_List(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockIAM := mock_iamiface.NewMockIAMAPI(ctrl)

	user := &iam.User{
		UserName: ptr.String("login-profile:foobar"),
	}

	mockIAM.EXPECT().ListUsers(nil).Return(&iam.ListUsersOutput{
		Users: []*iam.User{
			user,
		},
	}, nil)

	now := time.Now().UTC()

	mockIAM.EXPECT().GetLoginProfile(&iam.GetLoginProfileInput{
		UserName: user.UserName,
	}).Return(&iam.GetLoginProfileOutput{
		LoginProfile: &iam.LoginProfile{
			UserName:   user.UserName,
			CreateDate: ptr.Time(now),
		},
	}, nil)

	lister := IAMLoginProfileLister{
		mockSvc: mockIAM,
	}

	resources, err := lister.List(context.TODO(), testListerOpts)
	a.Nil(err)
	a.Len(resources, 1)
}

func Test_Mock_IAMLoginProfile_Remove(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockIAM := mock_iamiface.NewMockIAMAPI(ctrl)

	iamLoginProfile := IAMLoginProfile{
		svc:      mockIAM,
		UserName: ptr.String("login-profile:foobar"),
	}

	mockIAM.EXPECT().DeleteLoginProfile(gomock.Eq(&iam.DeleteLoginProfileInput{
		UserName: iamLoginProfile.UserName,
	})).Return(&iam.DeleteLoginProfileOutput{}, nil)

	err := iamLoginProfile.Remove(context.TODO())
	a.Nil(err)
}

func Test_Mock_IAMLoginProfile_Properties(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockIAM := mock_iamiface.NewMockIAMAPI(ctrl)

	iamLoginProfile := IAMLoginProfile{
		svc:      mockIAM,
		UserName: ptr.String("login-profile:foobar"),
	}

	a.Equal("login-profile:foobar", iamLoginProfile.Properties().Get("UserName"))
}
