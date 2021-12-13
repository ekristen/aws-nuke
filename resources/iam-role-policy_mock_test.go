package resources

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/golang/mock/gomock"
	"github.com/rebuy-de/aws-nuke/v2/mocks/mock_iamiface"
	"github.com/stretchr/testify/assert"
)

func Test_Mock_IAMRolePolicy_Remove(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockIAM := mock_iamiface.NewMockIAMAPI(ctrl)

	iamRolePolicy := IAMRolePolicy{
		svc:        mockIAM,
		roleId:     "role:foobar-id",
		roleName:   "role:foobar",
		policyName: "policy:foobar",
		roleTags:   []*iam.Tag{},
	}

	mockIAM.EXPECT().DeleteRolePolicy(gomock.Eq(&iam.DeleteRolePolicyInput{
		RoleName:   aws.String(iamRolePolicy.roleName),
		PolicyName: aws.String(iamRolePolicy.policyName),
	})).Return(&iam.DeleteRolePolicyOutput{}, nil)

	err := iamRolePolicy.Remove()
	a.Nil(err)
}
