package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"

	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const SageMakerUserProfilesResource = "SageMakerUserProfiles"

func init() {
	resource.Register(resource.Registration{
		Name:   SageMakerUserProfilesResource,
		Scope:  nuke.Account,
		Lister: &SageMakerUserProfilesLister{},
	})
}

type SageMakerUserProfilesLister struct{}

func (l *SageMakerUserProfilesLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := sagemaker.New(opts.Session)
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
			resources = append(resources, &SageMakerUserProfile{
				svc:             svc,
				domainID:        userProfile.DomainId,
				userProfileName: userProfile.UserProfileName,
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
	svc             *sagemaker.SageMaker
	domainID        *string
	userProfileName *string
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
	properties.
		Set("DomainID", f.domainID).
		Set("UserProfileName", f.userProfileName)
	return properties
}
