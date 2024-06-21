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

func Test_Mock_IAMRolePolicy_Remove(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockIAM := mock_iamiface.NewMockIAMAPI(ctrl)

	iamRolePolicy := IAMRolePolicy{
		svc:        mockIAM,
		roleID:     "role:foobar-id",
		roleName:   "role:foobar",
		policyName: "policy:foobar",
		roleTags:   []*iam.Tag{},
	}

	mockIAM.EXPECT().DeleteRolePolicy(gomock.Eq(&iam.DeleteRolePolicyInput{
		RoleName:   aws.String(iamRolePolicy.roleName),
		PolicyName: aws.String(iamRolePolicy.policyName),
	})).Return(&iam.DeleteRolePolicyOutput{}, nil)

	err := iamRolePolicy.Remove(context.TODO())
	a.Nil(err)
}
