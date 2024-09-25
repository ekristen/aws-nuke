package resources

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"

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
		Name:     ptr.String("foobar"),
		PolicyID: ptr.String("foobar"),
		ARN:      ptr.String("foobar"),
	}

	mockIAM.EXPECT().ListPolicyVersions(gomock.Eq(&iam.ListPolicyVersionsInput{
		PolicyArn: iamPolicy.ARN,
	})).Return(&iam.ListPolicyVersionsOutput{
		Versions: []*iam.PolicyVersion{
			{
				IsDefaultVersion: ptr.Bool(true),
				VersionId:        ptr.String("v1"),
			},
		},
	}, nil)

	mockIAM.EXPECT().DeletePolicy(gomock.Eq(&iam.DeletePolicyInput{
		PolicyArn: iamPolicy.ARN,
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
		Name:     ptr.String("foobar"),
		PolicyID: ptr.String("foobar"),
		ARN:      ptr.String("foobar"),
	}

	mockIAM.EXPECT().ListPolicyVersions(gomock.Eq(&iam.ListPolicyVersionsInput{
		PolicyArn: iamPolicy.ARN,
	})).Return(&iam.ListPolicyVersionsOutput{
		Versions: []*iam.PolicyVersion{
			{
				IsDefaultVersion: ptr.Bool(false),
				VersionId:        ptr.String("v1"),
			},
			{
				IsDefaultVersion: ptr.Bool(false),
				VersionId:        ptr.String("v2"),
			},
			{
				IsDefaultVersion: ptr.Bool(true),
				VersionId:        ptr.String("v3"),
			},
		},
	}, nil)

	mockIAM.EXPECT().DeletePolicyVersion(gomock.Eq(&iam.DeletePolicyVersionInput{
		PolicyArn: iamPolicy.ARN,
		VersionId: ptr.String("v1"),
	})).Return(&iam.DeletePolicyVersionOutput{}, nil)

	mockIAM.EXPECT().DeletePolicyVersion(gomock.Eq(&iam.DeletePolicyVersionInput{
		PolicyArn: iamPolicy.ARN,
		VersionId: ptr.String("v2"),
	})).Return(&iam.DeletePolicyVersionOutput{}, nil)

	mockIAM.EXPECT().DeletePolicy(gomock.Eq(&iam.DeletePolicyInput{
		PolicyArn: iamPolicy.ARN,
	})).Return(&iam.DeletePolicyOutput{}, nil)

	err := iamPolicy.Remove(context.TODO())
	a.Nil(err)
}

func Test_Mock_IAMPolicy_Properties(t *testing.T) {
	a := assert.New(t)

	now := time.Now().UTC()

	iamPolicy := IAMPolicy{
		Name:       ptr.String("foobar"),
		PolicyID:   ptr.String("foobar"),
		ARN:        ptr.String("arn:foobar"),
		Path:       ptr.String("/foobar"),
		CreateDate: ptr.Time(now),
		Tags: []*iam.Tag{
			{
				Key:   ptr.String("foo"),
				Value: ptr.String("bar"),
			},
		},
	}

	a.Equal("foobar", iamPolicy.Properties().Get("Name"))
	a.Equal("foobar", iamPolicy.Properties().Get("PolicyID"))
	a.Equal("arn:foobar", iamPolicy.Properties().Get("ARN"))
	a.Equal("/foobar", iamPolicy.Properties().Get("Path"))
	a.Equal("bar", iamPolicy.Properties().Get("tag:foo"))
	a.Equal(now.Format(time.RFC3339), iamPolicy.Properties().Get("CreateDate"))
	a.Equal("arn:foobar", iamPolicy.String())
}
