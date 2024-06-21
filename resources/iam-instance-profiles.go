package resources

import (
	"context"

	"github.com/gotidy/ptr"
	"github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const IAMInstanceProfileResource = "IAMInstanceProfile"

func init() {
	registry.Register(&registry.Registration{
		Name:   IAMInstanceProfileResource,
		Scope:  nuke.Account,
		Lister: &IAMInstanceProfileLister{},
		DeprecatedAliases: []string{
			"IamInstanceProfile",
		},
	})
}

type IAMInstanceProfileLister struct{}

func (l *IAMInstanceProfileLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := iam.New(opts.Session)
	params := &iam.ListInstanceProfilesInput{}
	resources := make([]resource.Resource, 0)

	for {
		resp, err := svc.ListInstanceProfiles(params)
		if err != nil {
			return nil, err
		}

		for _, out := range resp.InstanceProfiles {
			profile, err := GetIAMInstanceProfile(svc, out.InstanceProfileName)
			if err != nil {
				logrus.
					WithError(err).
					WithField("instanceProfileName", *out.InstanceProfileName).
					Error("Failed to get listed instance profile")
				continue
			}

			resources = append(resources, &IAMInstanceProfile{
				svc:     svc,
				name:    *out.InstanceProfileName,
				path:    *profile.Path,
				profile: profile,
			})
		}

		if !ptr.ToBool(resp.IsTruncated) {
			break
		}

		params.Marker = resp.Marker
	}

	return resources, nil
}

func GetIAMInstanceProfile(svc *iam.IAM, instanceProfileName *string) (*iam.InstanceProfile, error) {
	params := &iam.GetInstanceProfileInput{
		InstanceProfileName: instanceProfileName,
	}
	resp, err := svc.GetInstanceProfile(params)
	return resp.InstanceProfile, err
}

type IAMInstanceProfile struct {
	svc     iamiface.IAMAPI
	name    string
	path    string
	profile *iam.InstanceProfile
}

func (e *IAMInstanceProfile) Remove(_ context.Context) error {
	_, err := e.svc.DeleteInstanceProfile(&iam.DeleteInstanceProfileInput{
		InstanceProfileName: &e.name,
	})
	if err != nil {
		return err
	}

	return nil
}

func (e *IAMInstanceProfile) String() string {
	return e.name
}

func (e *IAMInstanceProfile) Properties() types.Properties {
	properties := types.NewProperties()

	for _, tagValue := range e.profile.Tags {
		properties.SetTag(tagValue.Key, tagValue.Value)
	}

	properties.
		Set("Name", e.name).
		Set("Path", e.path)

	return properties
}
