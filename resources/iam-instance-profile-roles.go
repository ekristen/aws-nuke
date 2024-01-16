package resources

import (
	"context"

	"fmt"
	"time"

	"github.com/gotidy/ptr"
	"github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"

	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const IAMInstanceProfileRoleResource = "IAMInstanceProfileRole"

func init() {
	resource.Register(resource.Registration{
		Name:   IAMInstanceProfileRoleResource,
		Scope:  nuke.Account,
		Lister: &IAMInstanceProfileRoleLister{},
	})
}

type IAMInstanceProfileRoleLister struct{}

func (l *IAMInstanceProfileRoleLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
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
			for _, outRole := range out.Roles {
				profile, err := GetIAMInstanceProfile(svc, out.InstanceProfileName)
				if err != nil {
					logrus.
						WithError(err).
						WithField("instanceProfileName", *out.InstanceProfileName).
						Error("Failed to get listed instance profile")
					continue
				}

				resources = append(resources, &IAMInstanceProfileRole{
					svc:     svc,
					role:    outRole,
					profile: profile,
				})
			}
		}

		if !*resp.IsTruncated {
			break
		}

		params.Marker = resp.Marker
	}

	return resources, nil
}

type IAMInstanceProfileRole struct {
	svc     iamiface.IAMAPI
	role    *iam.Role
	profile *iam.InstanceProfile
}

func (e *IAMInstanceProfileRole) Remove(_ context.Context) error {
	_, err := e.svc.RemoveRoleFromInstanceProfile(
		&iam.RemoveRoleFromInstanceProfileInput{
			InstanceProfileName: e.profile.InstanceProfileName,
			RoleName:            e.role.RoleName,
		})
	if err != nil {
		return err
	}

	return nil
}

func (e *IAMInstanceProfileRole) String() string {
	return fmt.Sprintf("%s -> %s", ptr.ToString(e.profile.InstanceProfileName), ptr.ToString(e.role.RoleName))
}

func (e *IAMInstanceProfileRole) Properties() types.Properties {
	properties := types.NewProperties()

	for _, tagValue := range e.profile.Tags {
		properties.SetTag(tagValue.Key, tagValue.Value)
	}

	properties.
		Set("InstanceProfile", e.profile.InstanceProfileName).
		Set("InstanceRole", e.role.RoleName).
		Set("role:Path", e.role.Path).
		Set("role:CreateDate", e.role.CreateDate.Format(time.RFC3339)).
		Set("role:LastUsedDate", getLastUsedDate(e.role, time.RFC3339))

	for _, tagValue := range e.role.Tags {
		properties.SetTagWithPrefix("role", tagValue.Key, tagValue.Value)
	}

	return properties
}
