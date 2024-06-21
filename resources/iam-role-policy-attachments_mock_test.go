package resources

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"

	"github.com/ekristen/aws-nuke/v3/mocks/mock_iamiface"
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
		role: &iam.Role{
			RoleName: aws.String("foo"),
		},
	}

	mockIAM.EXPECT().DetachRolePolicy(gomock.Eq(&iam.DetachRolePolicyInput{
		RoleName:  iamRolePolicyAttachment.role.RoleName,
		PolicyArn: aws.String(iamRolePolicyAttachment.policyArn),
	})).Return(&iam.DetachRolePolicyOutput{}, nil)

	err := iamRolePolicyAttachment.Remove(context.TODO())
	a.Nil(err)
}
