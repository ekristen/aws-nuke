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

func Test_Mock_IAMUserAccessKey_Remove(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockIAM := mock_iamiface.NewMockIAMAPI(ctrl)

	iamUserAccessKey := IAMUserAccessKey{
		svc:         mockIAM,
		accessKeyID: "EXAMPLEfoobar",
		userName:    "foobar",
		status:      "unknown",
	}

	mockIAM.EXPECT().DeleteAccessKey(gomock.Eq(&iam.DeleteAccessKeyInput{
		AccessKeyId: aws.String(iamUserAccessKey.accessKeyID),
		UserName:    aws.String(iamUserAccessKey.userName),
	})).Return(&iam.DeleteAccessKeyOutput{}, nil)

	err := iamUserAccessKey.Remove(context.TODO())
	a.Nil(err)
}
