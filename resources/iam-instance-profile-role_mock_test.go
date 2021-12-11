package resources

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/golang/mock/gomock"
	"github.com/rebuy-de/aws-nuke/mocks/mock_iamiface"
	"github.com/stretchr/testify/assert"
)

func Test_Mock_IAMInstanceProfileRole_Remove(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockIAM := mock_iamiface.NewMockIAMAPI(ctrl)

	iamInstanceProfileRole := IAMInstanceProfileRole{
		svc:     mockIAM,
		role:    "role:foobar",
		profile: "profile:foobar",
	}

	mockIAM.EXPECT().RemoveRoleFromInstanceProfile(gomock.Eq(&iam.RemoveRoleFromInstanceProfileInput{
		RoleName:            aws.String(iamInstanceProfileRole.role),
		InstanceProfileName: aws.String(iamInstanceProfileRole.profile),
	})).Return(&iam.RemoveRoleFromInstanceProfileOutput{}, nil)

	err := iamInstanceProfileRole.Remove()
	a.Nil(err)
}
