package resources

import (
	"context"
	"testing"
	"time"

	"github.com/gotidy/ptr"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"

	"github.com/ekristen/aws-nuke/v3/mocks/mock_iamiface"
	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

func Test_Mock_IAMRole_List(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockIAM := mock_iamiface.NewMockIAMAPI(ctrl)

	createDate := time.Now().Add(-24 * time.Hour).UTC()
	lastUsedDate := time.Now().Add(-8 * time.Hour).UTC()

	testRole := &iam.Role{
		Arn:          ptr.String("arn:aws:iam::123456789012:role/test"),
		RoleName:     ptr.String("test"),
		CreateDate:   ptr.Time(createDate),
		Path:         ptr.String("/"),
		RoleId:       ptr.String("test"),
		RoleLastUsed: &iam.RoleLastUsed{LastUsedDate: ptr.Time(lastUsedDate)},
	}

	mockIAM.EXPECT().ListRoles(gomock.Any()).Return(&iam.ListRolesOutput{
		Roles: []*iam.Role{
			testRole,
		},
		IsTruncated: ptr.Bool(false),
	}, nil)

	mockIAM.EXPECT().GetRole(&iam.GetRoleInput{
		RoleName: ptr.String("test"),
	}).Return(&iam.GetRoleOutput{
		Role: testRole,
	}, nil)

	lister := IAMRoleLister{
		mockSvc: mockIAM,
	}

	resources, err := lister.List(context.TODO(), &nuke.ListerOpts{
		Region: &nuke.Region{
			Name: "us-east-2",
		},
		Session: session.Must(session.NewSession()),
	})
	a.Nil(err)
	a.Len(resources, 1)

	iamRole := resources[0].(*IAMRole)
	a.Equal("test", *iamRole.Name)
	a.Equal("/", *iamRole.Path)
	a.Equal(createDate.Format(time.RFC3339), iamRole.Properties().Get("CreateDate"))
	a.Equal(lastUsedDate.Format(time.RFC3339), iamRole.Properties().Get("LastUsedDate"))
}

func Test_Mock_IAMRole_Remove(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockIAM := mock_iamiface.NewMockIAMAPI(ctrl)

	iamRole := IAMRole{
		svc:  mockIAM,
		Name: ptr.String("test"),
		Path: ptr.String("/"),
		Tags: []*iam.Tag{},
	}

	mockIAM.EXPECT().DeleteRole(gomock.Eq(&iam.DeleteRoleInput{
		RoleName: iamRole.Name,
	})).Return(&iam.DeleteRoleOutput{}, nil)

	err := iamRole.Remove(context.TODO())
	a.Nil(err)
}

func Test_Mock_IAMRole_Properties(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockIAM := mock_iamiface.NewMockIAMAPI(ctrl)

	iamRole := IAMRole{
		svc:  mockIAM,
		Name: ptr.String("test"),
		Path: ptr.String("/testing"),
		Tags: []*iam.Tag{
			{
				Key:   ptr.String("test-key"),
				Value: ptr.String("test"),
			},
		},
	}

	a.Equal("test", iamRole.Properties().Get("Name"))
	a.Equal("/testing", iamRole.Properties().Get("Path"))
	a.Equal("test", iamRole.Properties().Get("tag:test-key"))
}
