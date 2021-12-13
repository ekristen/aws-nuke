package resources

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/golang/mock/gomock"
	"github.com/rebuy-de/aws-nuke/mocks/mock_iamiface"
	"github.com/stretchr/testify/assert"
)

func Test_Mock_IAMGroupPolicyAttachment_Remove(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockIAM := mock_iamiface.NewMockIAMAPI(ctrl)

	iamGroupPolicyAttachment := IAMGroupPolicyAttachment{
		svc:        mockIAM,
		policyArn:  "foobar",
		policyName: "foobar",
		groupName:  "foobar",
	}

	mockIAM.EXPECT().DetachGroupPolicy(gomock.Eq(&iam.DetachGroupPolicyInput{
		PolicyArn: aws.String(iamGroupPolicyAttachment.policyArn),
		GroupName: aws.String(iamGroupPolicyAttachment.groupName),
	})).Return(&iam.DetachGroupPolicyOutput{}, nil)

	err := iamGroupPolicyAttachment.Remove()
	a.Nil(err)
}
