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

	err := iamGroupPolicyAttachment.Remove(context.TODO())
	a.Nil(err)
}
