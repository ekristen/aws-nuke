package resources

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/golang/mock/gomock"
	"github.com/rebuy-de/aws-nuke/mocks/mock_iamiface"
	"github.com/stretchr/testify/assert"
)

func Test_Mock_IAMServiceSpecificCredential_Remove(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockIAM := mock_iamiface.NewMockIAMAPI(ctrl)

	iamServiceSpecificCredential := IAMServiceSpecificCredential{
		svc:         mockIAM,
		name:        "user:foobar",
		serviceName: "service:foobar",
		id:          "user:service:foobar",
		userName:    "user:foobar",
	}

	mockIAM.EXPECT().DeleteServiceSpecificCredential(gomock.Eq(&iam.DeleteServiceSpecificCredentialInput{
		UserName:                    aws.String(iamServiceSpecificCredential.userName),
		ServiceSpecificCredentialId: aws.String(iamServiceSpecificCredential.id),
	})).Return(&iam.DeleteServiceSpecificCredentialOutput{}, nil)

	err := iamServiceSpecificCredential.Remove()
	a.Nil(err)
}
