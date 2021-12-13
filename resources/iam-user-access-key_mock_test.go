package resources

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/golang/mock/gomock"
	"github.com/rebuy-de/aws-nuke/mocks/mock_iamiface"
	"github.com/stretchr/testify/assert"
)

func Test_Mock_IAMUserAccessKey_Remove(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockIAM := mock_iamiface.NewMockIAMAPI(ctrl)

	iamUserAccessKey := IAMUserAccessKey{
		svc:         mockIAM,
		accessKeyId: "EXAMPLEfoobar",
		userName:    "foobar",
		status:      "unknown",
	}

	mockIAM.EXPECT().DeleteAccessKey(gomock.Eq(&iam.DeleteAccessKeyInput{
		AccessKeyId: aws.String(iamUserAccessKey.accessKeyId),
		UserName:    aws.String(iamUserAccessKey.userName),
	})).Return(&iam.DeleteAccessKeyOutput{}, nil)

	err := iamUserAccessKey.Remove()
	a.Nil(err)
}
