package resources

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/aws/aws-sdk-go/service/iam"

	"github.com/ekristen/aws-nuke/mocks/mock_iamiface"
)

func Test_Mock_IAMServerCertificate_Remove(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockIAM := mock_iamiface.NewMockIAMAPI(ctrl)

	iamServerCertificate := IAMServerCertificate{
		svc:  mockIAM,
		name: "server-cert-foobar",
	}

	mockIAM.EXPECT().DeleteServerCertificate(gomock.Eq(&iam.DeleteServerCertificateInput{
		ServerCertificateName: &iamServerCertificate.name,
	})).Return(&iam.DeleteServerCertificateOutput{}, nil)

	err := iamServerCertificate.Remove(context.TODO())
	a.Nil(err)
}
