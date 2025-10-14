package resources

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/aws/aws-sdk-go/aws"         //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/iam" //nolint:staticcheck

	"github.com/ekristen/aws-nuke/v3/mocks/mock_iamiface"
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

	err := iamUserGroup.Remove(context.TODO())
	a.Nil(err)
}
