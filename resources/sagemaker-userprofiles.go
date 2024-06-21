package resources

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"

	"github.com/aws/aws-sdk-go/service/sagemaker/sagemakeriface"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const SageMakerUserProfilesResource = "SageMakerUserProfiles"

func init() {
	registry.Register(&registry.Registration{
		Name:   SageMakerUserProfilesResource,
		Scope:  nuke.Account,
		Lister: &SageMakerUserProfilesLister{},
	})
}

type SageMakerUserProfilesLister struct {
	mockSvc sagemakeriface.SageMakerAPI
}

func (l *SageMakerUserProfilesLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	var svc sagemakeriface.SageMakerAPI
	if l.mockSvc != nil {
		svc = l.mockSvc
	} else {
		svc = sagemaker.New(opts.Session)
	}

	resources := make([]resource.Resource, 0)

	params := &sagemaker.ListUserProfilesInput{
		MaxResults: aws.Int64(30),
	}

	for {
		resp, err := svc.ListUserProfiles(params)
		if err != nil {
			return nil, err
		}

		for _, userProfile := range resp.UserProfiles {
			var tags []*sagemaker.Tag
			up, err := svc.DescribeUserProfile(&sagemaker.DescribeUserProfileInput{
				DomainId:        userProfile.DomainId,
				UserProfileName: userProfile.UserProfileName,
			})
			if err != nil {
				logrus.WithError(err).Error("unable to get user profile")
				continue
			}

			upTags, err := svc.ListTags(&sagemaker.ListTagsInput{
				ResourceArn: up.UserProfileArn,
				MaxResults:  aws.Int64(100),
			})
			if err != nil {
				logrus.WithError(err).Error("unable to get tags")
				continue
			}
			if upTags.Tags != nil {
				tags = upTags.Tags
			}

			resources = append(resources, &SageMakerUserProfile{
				svc:             svc,
				domainID:        userProfile.DomainId,
				userProfileName: userProfile.UserProfileName,
				tags:            tags,
			})
		}

		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return resources, nil
}

type SageMakerUserProfile struct {
	svc             sagemakeriface.SageMakerAPI
	domainID        *string
	userProfileName *string
	tags            []*sagemaker.Tag
}

func (f *SageMakerUserProfile) Remove(_ context.Context) error {
	_, err := f.svc.DeleteUserProfile(&sagemaker.DeleteUserProfileInput{
		DomainId:        f.domainID,
		UserProfileName: f.userProfileName,
	})

	return err
}

func (f *SageMakerUserProfile) String() string {
	return *f.userProfileName
}

func (f *SageMakerUserProfile) Properties() types.Properties {
	properties := types.NewProperties()

	properties.Set("DomainID", f.domainID)
	properties.Set("UserProfileName", f.userProfileName)

	for _, tag := range f.tags {
		properties.SetTag(tag.Key, tag.Value)
	}

	return properties
}
