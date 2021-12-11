package resources

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/golang/mock/gomock"
	"github.com/rebuy-de/aws-nuke/mocks/mock_iamiface"
	"github.com/stretchr/testify/assert"
)

func Test_Mock_IAMUserSSHKeys_Remove(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockIAM := mock_iamiface.NewMockIAMAPI(ctrl)

	iamUserSSHKey := IAMUserSSHKey{
		svc:      mockIAM,
		userName: "foobar",
		sshKeyID: "foobar",
	}

	mockIAM.EXPECT().DeleteSSHPublicKey(gomock.Eq(&iam.DeleteSSHPublicKeyInput{
		UserName:       aws.String(iamUserSSHKey.userName),
		SSHPublicKeyId: aws.String(iamUserSSHKey.sshKeyID),
	})).Return(&iam.DeleteSSHPublicKeyOutput{}, nil)

	err := iamUserSSHKey.Remove()
	a.Nil(err)
}
