package resources

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/golang/mock/gomock"
	"github.com/rebuy-de/aws-nuke/mocks/mock_iamiface"
	"github.com/stretchr/testify/assert"
)

func Test_Mock_IAMRolePolicyAttachment_Remove(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockIAM := mock_iamiface.NewMockIAMAPI(ctrl)

	iamRolePolicyAttachment := IAMRolePolicyAttachment{
		svc:        mockIAM,
		policyArn:  "arn:foobar",
		policyName: "foobar",
		roleName:   "role:foobar",
	}

	mockIAM.EXPECT().DetachRolePolicy(gomock.Eq(&iam.DetachRolePolicyInput{
		RoleName:  aws.String(iamRolePolicyAttachment.roleName),
		PolicyArn: aws.String(iamRolePolicyAttachment.policyArn),
	})).Return(&iam.DetachRolePolicyOutput{}, nil)

	err := iamRolePolicyAttachment.Remove()
	a.Nil(err)
}
