package resources

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/aws/aws-sdk-go/service/iam"

	"github.com/ekristen/aws-nuke/mocks/mock_iamiface"
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

	err := iamSAMLProvider.Remove(context.TODO())
	a.Nil(err)
}
