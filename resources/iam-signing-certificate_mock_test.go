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

func Test_Mock_IAMSigningCertificate_Remove(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockIAM := mock_iamiface.NewMockIAMAPI(ctrl)

	iamSigningCertificate := IAMSigningCertificate{
		svc:           mockIAM,
		certificateID: aws.String("certid:foobar"),
		userName:      aws.String("user:foobar"),
		status:        aws.String("unknown"),
	}

	mockIAM.EXPECT().DeleteSigningCertificate(gomock.Eq(&iam.DeleteSigningCertificateInput{
		UserName:      iamSigningCertificate.userName,
		CertificateId: iamSigningCertificate.certificateID,
	})).Return(&iam.DeleteSigningCertificateOutput{}, nil)

	err := iamSigningCertificate.Remove(context.TODO())
	a.Nil(err)
}
