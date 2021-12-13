package resources

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/golang/mock/gomock"
	"github.com/rebuy-de/aws-nuke/v2/mocks/mock_iamiface"
	"github.com/stretchr/testify/assert"
)

func Test_Mock_IAMRole_Remove(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockIAM := mock_iamiface.NewMockIAMAPI(ctrl)

	iamRole := IAMRole{
		svc:  mockIAM,
		name: "test",
		path: "/",
		tags: []*iam.Tag{},
	}

	mockIAM.EXPECT().DeleteRole(gomock.Eq(&iam.DeleteRoleInput{
		RoleName: aws.String(iamRole.name),
	})).Return(&iam.DeleteRoleOutput{}, nil)

	err := iamRole.Remove()
	a.Nil(err)
}
