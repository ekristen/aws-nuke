package resources

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"

	"github.com/aws/aws-sdk-go/service/iam" //nolint:staticcheck

	"github.com/ekristen/aws-nuke/v3/mocks/mock_iamiface"
)

func Test_Mock_IAMServiceSpecificCredential_Remove(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockIAM := mock_iamiface.NewMockIAMAPI(ctrl)

	iamServiceSpecificCredential := IAMServiceSpecificCredential{
		svc:         mockIAM,
		name:        ptr.String("user:foobar"),
		serviceName: ptr.String("service:foobar"),
		id:          ptr.String("user:service:foobar"),
		userName:    ptr.String("user:foobar"),
	}

	mockIAM.EXPECT().DeleteServiceSpecificCredential(gomock.Eq(&iam.DeleteServiceSpecificCredentialInput{
		UserName:                    iamServiceSpecificCredential.userName,
		ServiceSpecificCredentialId: iamServiceSpecificCredential.id,
	})).Return(&iam.DeleteServiceSpecificCredentialOutput{}, nil)

	err := iamServiceSpecificCredential.Remove(context.TODO())
	a.Nil(err)
}
