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

func Test_Mock_IAMUserPolicy_Remove(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockIAM := mock_iamiface.NewMockIAMAPI(ctrl)

	iamUserPolicy := IAMUserPolicy{
		svc:        mockIAM,
		userName:   "foobar",
		policyName: "foobar",
	}

	mockIAM.EXPECT().DeleteUserPolicy(gomock.Eq(&iam.DeleteUserPolicyInput{
		UserName:   aws.String(iamUserPolicy.userName),
		PolicyName: aws.String(iamUserPolicy.policyName),
	})).Return(&iam.DeleteUserPolicyOutput{}, nil)

	err := iamUserPolicy.Remove(context.TODO())
	a.Nil(err)
}
