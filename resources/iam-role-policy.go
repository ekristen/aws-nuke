package resources

import (
	"context"

	"fmt"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const IAMRolePolicyResource = "IAMRolePolicy"

func init() {
	registry.Register(&registry.Registration{
		Name:   IAMRolePolicyResource,
		Scope:  nuke.Account,
		Lister: &IAMRolePolicyLister{},
	})
}

type IAMRolePolicy struct {
	svc        iamiface.IAMAPI
	roleID     string
	rolePath   string
	roleName   string
	policyName string
	roleTags   []*iam.Tag
}

func (e *IAMRolePolicy) Filter() error {
	if strings.HasPrefix(e.rolePath, "/aws-service-role/") {
		return fmt.Errorf("cannot alter service roles")
	}
	return nil
}

func (e *IAMRolePolicy) Remove(_ context.Context) error {
	_, err := e.svc.DeleteRolePolicy(
		&iam.DeleteRolePolicyInput{
			RoleName:   &e.roleName,
			PolicyName: &e.policyName,
		})
	if err != nil {
		return err
	}

	return nil
}

func (e *IAMRolePolicy) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("PolicyName", e.policyName)
	properties.Set("role:RoleName", e.roleName)
	properties.Set("role:RoleID", e.roleID)
	properties.Set("role:Path", e.rolePath)

	for _, tagValue := range e.roleTags {
		properties.SetTagWithPrefix("role", tagValue.Key, tagValue.Value)
	}
	return properties
}

func (e *IAMRolePolicy) String() string {
	return fmt.Sprintf("%s -> %s", e.roleName, e.policyName)
}

// ----------------------

type IAMRolePolicyLister struct{}

func (l *IAMRolePolicyLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := iam.New(opts.Session)
	roleParams := &iam.ListRolesInput{}
	resources := make([]resource.Resource, 0)

	for {
		roles, err := svc.ListRoles(roleParams)
		if err != nil {
			return nil, err
		}

		for _, listedRole := range roles.Roles {
			role, err := GetIAMRole(svc, listedRole.RoleName)
			if err != nil {
				logrus.Errorf("Failed to get listed role %s: %v", *listedRole.RoleName, err)
				continue
			}

			polParams := &iam.ListRolePoliciesInput{
				RoleName: role.RoleName,
			}

			for {
				policies, err := svc.ListRolePolicies(polParams)
				if err != nil {
					logrus.
						WithError(err).
						WithField("roleName", *role.RoleName).
						Error("Failed to list policies")
					break
				}

				for _, policyName := range policies.PolicyNames {
					resources = append(resources, &IAMRolePolicy{
						svc:        svc,
						roleID:     *role.RoleId,
						roleName:   *role.RoleName,
						rolePath:   *role.Path,
						policyName: *policyName,
						roleTags:   role.Tags,
					})
				}

				if !*policies.IsTruncated {
					break
				}

				polParams.Marker = policies.Marker
			}
		}

		if !*roles.IsTruncated {
			break
		}

		roleParams.Marker = roles.Marker
	}

	return resources, nil
}
