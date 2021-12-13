package resources

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/golang/mock/gomock"
	"github.com/rebuy-de/aws-nuke/mocks/mock_iamiface"
	"github.com/stretchr/testify/assert"
)

func Test_Mock_IAMUserGroup_Remove(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockIAM := mock_iamiface.NewMockIAMAPI(ctrl)

	iamUserGroup := IAMUserGroupAttachment{
		svc:       mockIAM,
		userName:  "user:foobar",
		groupName: "group:foobar",
	}

	mockIAM.EXPECT().RemoveUserFromGroup(gomock.Eq(&iam.RemoveUserFromGroupInput{
		UserName:  aws.String(iamUserGroup.userName),
		GroupName: aws.String(iamUserGroup.groupName),
	})).Return(&iam.RemoveUserFromGroupOutput{}, nil)

	err := iamUserGroup.Remove()
	a.Nil(err)
}
