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

	err := iamUserPolicyAttachment.Remove(context.TODO())
	a.Nil(err)
}
