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

type IAMInstanceProfileLister struct {
	mockSvc iamiface.IAMAPI
}

func (l *IAMInstanceProfileLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	resources := make([]resource.Resource, 0)

	var svc iamiface.IAMAPI
	if l.mockSvc != nil {
		svc = l.mockSvc
	} else {
		svc = iam.New(opts.Session)
	}

	params := &iam.ListInstanceProfilesInput{}

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
				svc:  svc,
				Name: out.InstanceProfileName,
				Path: profile.Path,
				Tags: profile.Tags,
			})
		}

		if !ptr.ToBool(resp.IsTruncated) {
			break
		}

		params.Marker = resp.Marker
	}

	return resources, nil
}

type IAMInstanceProfile struct {
	svc  iamiface.IAMAPI
	Name *string
	Path *string
	Tags []*iam.Tag
}

func (r *IAMInstanceProfile) Remove(_ context.Context) error {
	_, err := r.svc.DeleteInstanceProfile(&iam.DeleteInstanceProfileInput{
		InstanceProfileName: r.Name,
	})
	if err != nil {
		return err
	}

	return nil
}

func (r *IAMInstanceProfile) String() string {
	return *r.Name
}

func (r *IAMInstanceProfile) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

// GetIAMInstanceProfile returns an IAM instance profile
func GetIAMInstanceProfile(svc iamiface.IAMAPI, instanceProfileName *string) (*iam.InstanceProfile, error) {
	resp, err := svc.GetInstanceProfile(&iam.GetInstanceProfileInput{
		InstanceProfileName: instanceProfileName,
	})
	if err != nil {
		return nil, err
	}

	return resp.InstanceProfile, nil
}
