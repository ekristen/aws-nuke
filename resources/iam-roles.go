package resources

import (
	"context"

	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"

	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const IAMRoleResource = "IAMRole"

func init() {
	resource.Register(&resource.Registration{
		Name:   IAMRoleResource,
		Scope:  nuke.Account,
		Lister: &IAMRoleLister{},
		DependsOn: []string{
			IAMRolePolicyAttachmentResource,
		},
		DeprecatedAliases: []string{
			"IamRole",
		},
	})
}

type IAMRole struct {
	svc  iamiface.IAMAPI
	name string
	path string
	tags []*iam.Tag
}

func (e *IAMRole) Filter() error {
	if strings.HasPrefix(e.path, "/aws-service-role/") {
		return fmt.Errorf("cannot delete service roles")
	}
	if strings.HasPrefix(e.path, "/aws-reserved/sso.amazonaws.com/") {
		return fmt.Errorf("cannot delete SSO roles")
	}
	return nil
}

func (e *IAMRole) Remove(_ context.Context) error {
	_, err := e.svc.DeleteRole(&iam.DeleteRoleInput{
		RoleName: aws.String(e.name),
	})
	if err != nil {
		return err
	}

	return nil
}

func (e *IAMRole) Properties() types.Properties {
	properties := types.NewProperties()
	for _, tagValue := range e.tags {
		properties.SetTag(tagValue.Key, tagValue.Value)
	}

	return properties
}

func (e *IAMRole) String() string {
	return e.name
}

// ---------

func GetIAMRole(svc *iam.IAM, roleName *string) (*iam.Role, error) {
	params := &iam.GetRoleInput{
		RoleName: roleName,
	}
	resp, err := svc.GetRole(params)
	return resp.Role, err
}

func getLastUsedDate(role *iam.Role, format string) string {
	var lastUsedDate *time.Time
	if role.RoleLastUsed == nil || role.RoleLastUsed.LastUsedDate == nil {
		lastUsedDate = role.CreateDate
	} else {
		lastUsedDate = role.RoleLastUsed.LastUsedDate
	}

	return lastUsedDate.Format(format)
}

// --------------

type IAMRoleLister struct{}

func (l *IAMRoleLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := iam.New(opts.Session)
	params := &iam.ListRolesInput{}
	resources := make([]resource.Resource, 0)

	for {
		resp, err := svc.ListRoles(params)
		if err != nil {
			return nil, err
		}

		for _, out := range resp.Roles {
			role, err := GetIAMRole(svc, out.RoleName)
			if err != nil {
				logrus.
					WithError(err).
					WithField("roleName", *out.RoleName).
					Error("Failed to get listed role")
				continue
			}

			resources = append(resources, &IAMRole{
				svc:  svc,
				name: *role.RoleName,
				path: *role.Path,
				tags: role.Tags,
			})
		}

		if !*resp.IsTruncated {
			break
		}

		params.Marker = resp.Marker
	}

	return resources, nil
}
