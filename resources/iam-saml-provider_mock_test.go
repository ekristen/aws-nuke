package resources

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/golang/mock/gomock"
	"github.com/rebuy-de/aws-nuke/v2/mocks/mock_iamiface"
	"github.com/stretchr/testify/assert"
)

func Test_Mock_IAMSAMLProvider_Remove(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockIAM := mock_iamiface.NewMockIAMAPI(ctrl)

	iamSAMLProvider := IAMSAMLProvider{
		svc: mockIAM,
		arn: "arn:mock-saml-provider-foobar",
	}

	mockIAM.EXPECT().DeleteSAMLProvider(gomock.Eq(&iam.DeleteSAMLProviderInput{
		SAMLProviderArn: &iamSAMLProvider.arn,
	})).Return(&iam.DeleteSAMLProviderOutput{}, nil)

	err := iamSAMLProvider.Remove()
	a.Nil(err)
}
