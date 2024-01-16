package resources

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"

	"github.com/ekristen/aws-nuke/mocks/mock_iamiface"
)

func Test_Mock_IAMInstanceProfileRole_Remove(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockIAM := mock_iamiface.NewMockIAMAPI(ctrl)

	iamInstanceProfileRole := IAMInstanceProfileRole{
		svc: mockIAM,
		role: &iam.Role{
			RoleName: aws.String("role:foobar"),
		},
		profile: &iam.InstanceProfile{
			InstanceProfileName: aws.String("profile:foobar"),
		},
	}

	mockIAM.EXPECT().RemoveRoleFromInstanceProfile(gomock.Eq(&iam.RemoveRoleFromInstanceProfileInput{
		RoleName:            iamInstanceProfileRole.role.RoleName,
		InstanceProfileName: iamInstanceProfileRole.profile.InstanceProfileName,
	})).Return(&iam.RemoveRoleFromInstanceProfileOutput{}, nil)

	err := iamInstanceProfileRole.Remove(context.TODO())
	a.Nil(err)
}
