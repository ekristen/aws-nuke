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

	err := iamGroupPolicy.Remove(context.TODO())
	a.Nil(err)
}
