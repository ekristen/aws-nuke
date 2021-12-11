package resources

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/golang/mock/gomock"
	"github.com/rebuy-de/aws-nuke/mocks/mock_iamiface"
	"github.com/stretchr/testify/assert"
)

func Test_Mock_IAMPolicy_Remove(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockIAM := mock_iamiface.NewMockIAMAPI(ctrl)

	iamPolicy := IAMPolicy{
		svc:      mockIAM,
		name:     "foobar",
		policyId: "foobar",
		arn:      "foobar",
	}

	mockIAM.EXPECT().ListPolicyVersions(gomock.Eq(&iam.ListPolicyVersionsInput{
		PolicyArn: aws.String(iamPolicy.arn),
	})).Return(&iam.ListPolicyVersionsOutput{
		Versions: []*iam.PolicyVersion{
			{
				IsDefaultVersion: aws.Bool(true),
				VersionId:        aws.String("v1"),
			},
		},
	}, nil)

	mockIAM.EXPECT().DeletePolicy(gomock.Eq(&iam.DeletePolicyInput{
		PolicyArn: aws.String(iamPolicy.arn),
	})).Return(&iam.DeletePolicyOutput{}, nil)

	err := iamPolicy.Remove()
	a.Nil(err)
}

func Test_Mock_IAMPolicy_WithVersions_Remove(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockIAM := mock_iamiface.NewMockIAMAPI(ctrl)

	iamPolicy := IAMPolicy{
		svc:      mockIAM,
		name:     "foobar",
		policyId: "foobar",
		arn:      "foobar",
	}

	mockIAM.EXPECT().ListPolicyVersions(gomock.Eq(&iam.ListPolicyVersionsInput{
		PolicyArn: aws.String(iamPolicy.arn),
	})).Return(&iam.ListPolicyVersionsOutput{
		Versions: []*iam.PolicyVersion{
			{
				IsDefaultVersion: aws.Bool(false),
				VersionId:        aws.String("v1"),
			},
			{
				IsDefaultVersion: aws.Bool(false),
				VersionId:        aws.String("v2"),
			},
			{
				IsDefaultVersion: aws.Bool(true),
				VersionId:        aws.String("v3"),
			},
		},
	}, nil)

	mockIAM.EXPECT().DeletePolicyVersion(gomock.Eq(&iam.DeletePolicyVersionInput{
		PolicyArn: aws.String(iamPolicy.arn),
		VersionId: aws.String("v1"),
	})).Return(&iam.DeletePolicyVersionOutput{}, nil)

	mockIAM.EXPECT().DeletePolicyVersion(gomock.Eq(&iam.DeletePolicyVersionInput{
		PolicyArn: aws.String(iamPolicy.arn),
		VersionId: aws.String("v2"),
	})).Return(&iam.DeletePolicyVersionOutput{}, nil)

	mockIAM.EXPECT().DeletePolicy(gomock.Eq(&iam.DeletePolicyInput{
		PolicyArn: aws.String(iamPolicy.arn),
	})).Return(&iam.DeletePolicyOutput{}, nil)

	err := iamPolicy.Remove()
	a.Nil(err)
}
