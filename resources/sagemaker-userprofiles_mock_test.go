package resources

import (
	"context"

	"testing"

	"github.com/golang/mock/gomock"
	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"

	"github.com/aws/aws-sdk-go/service/sagemaker" //nolint:staticcheck

	"github.com/ekristen/aws-nuke/v3/mocks/mock_sagemakeriface"
)

func Test_Mock_SageMakerUserProfiles_List(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mock_sagemakeriface.NewMockSageMakerAPI(ctrl)

	resource := SageMakerUserProfilesLister{
		mockSvc: mockSvc,
	}

	mockSvc.EXPECT().ListUserProfiles(gomock.Any()).Return(&sagemaker.ListUserProfilesOutput{
		UserProfiles: []*sagemaker.UserProfileDetails{
			{
				DomainId:        ptr.String("foo"),
				UserProfileName: ptr.String("bar"),
			},
		},
	}, nil)

	mockSvc.EXPECT().DescribeUserProfile(gomock.Eq(&sagemaker.DescribeUserProfileInput{
		DomainId:        ptr.String("foo"),
		UserProfileName: ptr.String("bar"),
	})).Return(&sagemaker.DescribeUserProfileOutput{
		DomainId:        ptr.String("foo"),
		UserProfileName: ptr.String("bar"),
		UserProfileArn:  ptr.String("arn:foobar"),
	}, nil)

	mockSvc.EXPECT().ListTags(gomock.Eq(&sagemaker.ListTagsInput{
		ResourceArn: ptr.String("arn:foobar"),
		MaxResults:  ptr.Int64(100),
	})).Return(&sagemaker.ListTagsOutput{
		Tags: []*sagemaker.Tag{
			{
				Key:   ptr.String("foo"),
				Value: ptr.String("bar"),
			},
		},
	}, nil)

	resources, err := resource.List(context.TODO(), testListerOpts)
	a.Nil(err)
	a.Len(resources, 1)

	userProfile := resources[0].(*SageMakerUserProfile)
	a.Equal("foo", *userProfile.domainID)
	a.Equal("bar", *userProfile.userProfileName)
}

func Test_Mock_SageMakerUserProfile_Remove(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mock_sagemakeriface.NewMockSageMakerAPI(ctrl)

	resource := SageMakerUserProfile{
		svc:             mockSvc,
		domainID:        ptr.String("foo"),
		userProfileName: ptr.String("bar"),
	}

	mockSvc.EXPECT().DeleteUserProfile(gomock.Eq(&sagemaker.DeleteUserProfileInput{
		DomainId:        resource.domainID,
		UserProfileName: resource.userProfileName,
	})).Return(&sagemaker.DeleteUserProfileOutput{}, nil)

	err := resource.Remove(context.TODO())
	a.Nil(err)
}
