package resources

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/golang/mock/gomock"
	"github.com/rebuy-de/aws-nuke/mocks/mock_iamiface"
	"github.com/stretchr/testify/assert"
)

func Test_Mock_IAMGroupPolicy_Remove(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockIAM := mock_iamiface.NewMockIAMAPI(ctrl)

	iamGroupPolicy := IAMGroupPolicy{
		svc:        mockIAM,
		policyName: "foobar",
		groupName:  "foobar",
	}

	mockIAM.EXPECT().DeleteGroupPolicy(gomock.Eq(&iam.DeleteGroupPolicyInput{
		PolicyName: aws.String(iamGroupPolicy.policyName),
		GroupName:  aws.String(iamGroupPolicy.groupName),
	})).Return(&iam.DeleteGroupPolicyOutput{}, nil)

	err := iamGroupPolicy.Remove()
	a.Nil(err)
}
