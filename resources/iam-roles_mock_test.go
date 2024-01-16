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

	err := iamRole.Remove(context.TODO())
	a.Nil(err)
}
