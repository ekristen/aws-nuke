package resources

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"

	"github.com/ekristen/aws-nuke/mocks/mock_iamiface"
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

	err := iamOpenIDConnectProvider.Remove(context.TODO())
	a.Nil(err)
}
