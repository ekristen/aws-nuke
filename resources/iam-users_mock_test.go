package resources

import (
	"context"

	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"

	"github.com/ekristen/aws-nuke/v3/mocks/mock_iamiface"
)

func Test_Mock_IAMUser_Remove(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockIAM := mock_iamiface.NewMockIAMAPI(ctrl)

	iamUser := IAMUser{
		svc:  mockIAM,
		name: "foobar",
	}

	mockIAM.EXPECT().DeleteUser(gomock.Eq(&iam.DeleteUserInput{
		UserName: aws.String(iamUser.name),
	})).Return(&iam.DeleteUserOutput{}, nil)

	err := iamUser.Remove(context.TODO())
	a.Nil(err)
}
