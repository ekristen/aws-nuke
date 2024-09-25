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

func Test_Mock_IAMPolicy_Remove(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockIAM := mock_iamiface.NewMockIAMAPI(ctrl)

	iamPolicy := IAMPolicy{
		svc:      mockIAM,
		name:     "foobar",
		policyID: "foobar",
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

	err := iamPolicy.Remove(context.TODO())
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
		policyID: "foobar",
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

	err := iamPolicy.Remove(context.TODO())
	a.Nil(err)
}

func Test_Mock_IAMPolicy_Properties(t *testing.T) {
	a := assert.New(t)

	iamPolicy := IAMPolicy{
		name:     "foobar",
		policyID: "foobar",
		arn:      "arn:foobar",
		path:     "/foobar",
		tags: []*iam.Tag{
			{
				Key:   aws.String("foo"),
				Value: aws.String("bar"),
			},
		},
	}

	a.Equal("foobar", iamPolicy.Properties().Get("Name"))
	a.Equal("foobar", iamPolicy.Properties().Get("PolicyID"))
	a.Equal("arn:foobar", iamPolicy.Properties().Get("ARN"))
	a.Equal("/foobar", iamPolicy.Properties().Get("Path"))
	a.Equal("bar", iamPolicy.Properties().Get("tag:foo"))
	a.Equal("arn:foobar", iamPolicy.String())
}
