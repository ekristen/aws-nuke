package resources

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/golang/mock/gomock"
	"github.com/rebuy-de/aws-nuke/v2/mocks/mock_iamiface"
	"github.com/stretchr/testify/assert"
)

func Test_Mock_IAMSigningCertificate_Remove(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockIAM := mock_iamiface.NewMockIAMAPI(ctrl)

	iamSigningCertificate := IAMSigningCertificate{
		svc:           mockIAM,
		certificateId: aws.String("certid:foobar"),
		userName:      aws.String("user:foobar"),
		status:        aws.String("unknown"),
	}

	mockIAM.EXPECT().DeleteSigningCertificate(gomock.Eq(&iam.DeleteSigningCertificateInput{
		UserName:      iamSigningCertificate.userName,
		CertificateId: iamSigningCertificate.certificateId,
	})).Return(&iam.DeleteSigningCertificateOutput{}, nil)

	err := iamSigningCertificate.Remove()
	a.Nil(err)
}
