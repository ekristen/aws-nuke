package resources

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/golang/mock/gomock"
	"github.com/rebuy-de/aws-nuke/mocks/mock_iamiface"
	"github.com/stretchr/testify/assert"
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

	err := iamUserPolicy.Remove()
	a.Nil(err)
}
