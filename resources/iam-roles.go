package resources

import (
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"
	"github.com/rebuy-de/aws-nuke/pkg/types"
	"github.com/sirupsen/logrus"
)

type IAMRole struct {
	svc  iamiface.IAMAPI
	name string
	path string
	tags []*iam.Tag
}

func init() {
	register("IAMRole", ListIAMRoles)
}

func GetIAMRole(svc *iam.IAM, roleName *string) (*iam.Role, error) {
	params := &iam.GetRoleInput{
		RoleName: roleName,
	}
	resp, err := svc.GetRole(params)
	return resp.Role, err
}

func ListIAMRoles(sess *session.Session) ([]Resource, error) {
	svc := iam.New(sess)
	params := &iam.ListRolesInput{}
	resources := make([]Resource, 0)

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

func (e *IAMRole) Filter() error {
	if strings.HasPrefix(e.path, "/aws-service-role/") {
		return fmt.Errorf("cannot delete service roles")
	}
	if strings.HasPrefix(e.path, "/aws-reserved/sso.amazonaws.com/") {
		return fmt.Errorf("cannot delete SSO roles")
	}
	return nil
}

func (e *IAMRole) Remove() error {
	_, err := e.svc.DeleteRole(&iam.DeleteRoleInput{
		RoleName: aws.String(e.name),
	})
	if err != nil {
		return err
	}

	return nil
}

func (role *IAMRole) Properties() types.Properties {
	properties := types.NewProperties()
	for _, tagValue := range role.tags {
		properties.SetTag(tagValue.Key, tagValue.Value)
	}

	return properties
}

func (e *IAMRole) String() string {
	return e.name
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
