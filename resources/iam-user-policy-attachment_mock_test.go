package resources

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/golang/mock/gomock"
	"github.com/rebuy-de/aws-nuke/mocks/mock_iamiface"
	"github.com/stretchr/testify/assert"
)

func Test_Mock_IAMUserPolicyAttachment_Remove(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockIAM := mock_iamiface.NewMockIAMAPI(ctrl)

	iamUserPolicyAttachment := IAMUserPolicyAttachment{
		svc:        mockIAM,
		policyArn:  "arn:foobar",
		policyName: "foobar",
		userName:   "foobar",
	}

	mockIAM.EXPECT().DetachUserPolicy(gomock.Eq(&iam.DetachUserPolicyInput{
		UserName:  aws.String(iamUserPolicyAttachment.userName),
		PolicyArn: aws.String(iamUserPolicyAttachment.policyArn),
	})).Return(&iam.DetachUserPolicyOutput{}, nil)

	err := iamUserPolicyAttachment.Remove()
	a.Nil(err)
}
