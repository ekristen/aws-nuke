package resources

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/golang/mock/gomock"
	"github.com/rebuy-de/aws-nuke/v2/mocks/mock_iamiface"
	"github.com/stretchr/testify/assert"
)

func Test_Mock_IAMOpenIDConnectProvider_Remove(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockIAM := mock_iamiface.NewMockIAMAPI(ctrl)

	iamOpenIDConnectProvider := IAMOpenIDConnectProvider{
		svc: mockIAM,
		arn: "arn:openid-connect-provider",
	}

	mockIAM.EXPECT().DeleteOpenIDConnectProvider(gomock.Eq(&iam.DeleteOpenIDConnectProviderInput{
		OpenIDConnectProviderArn: aws.String(iamOpenIDConnectProvider.arn),
	})).Return(&iam.DeleteOpenIDConnectProviderOutput{}, nil)

	err := iamOpenIDConnectProvider.Remove()
	a.Nil(err)
}
