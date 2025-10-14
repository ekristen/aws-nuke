package resources

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/service/iam" //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/iam/iamiface"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const IAMRolePolicyAttachmentResource = "IAMRolePolicyAttachment"

func init() {
	registry.Register(&registry.Registration{
		Name:     IAMRolePolicyAttachmentResource,
		Scope:    nuke.Account,
		Resource: &IAMRolePolicyAttachment{},
		Lister:   &IAMRolePolicyAttachmentLister{},
		DeprecatedAliases: []string{
			"IamRolePolicyAttachement",
		},
	})
}

type IAMRolePolicyAttachment struct {
	svc        iamiface.IAMAPI
	policyArn  string
	policyName string
	role       *iam.Role
}

func (e *IAMRolePolicyAttachment) Filter() error {
	if strings.Contains(e.policyArn, ":iam::aws:policy/aws-service-role/") {
		return fmt.Errorf("cannot detach from service roles")
	}
	if strings.HasPrefix(*e.role.Path, "/aws-reserved/sso.amazonaws.com/") {
		return fmt.Errorf("cannot detach from SSO roles")
	}
	return nil
}

func (e *IAMRolePolicyAttachment) Remove(_ context.Context) error {
	_, err := e.svc.DetachRolePolicy(
		&iam.DetachRolePolicyInput{
			PolicyArn: &e.policyArn,
			RoleName:  e.role.RoleName,
		})
	if err != nil {
		return err
	}

	return nil
}

func (e *IAMRolePolicyAttachment) Properties() types.Properties {
	properties := types.NewProperties().
		Set("RoleName", e.role.RoleName).
		Set("RolePath", e.role.Path).
		Set("RoleLastUsed", getLastUsedDate(e.role)).
		Set("RoleCreateDate", e.role.CreateDate.Format(time.RFC3339)).
		Set("PolicyName", e.policyName).
		Set("PolicyArn", e.policyArn)

	for _, tag := range e.role.Tags {
		properties.SetTagWithPrefix("role", tag.Key, tag.Value)
	}
	return properties
}

func (e *IAMRolePolicyAttachment) String() string {
	return fmt.Sprintf("%s -> %s", *e.role.RoleName, e.policyName)
}

// -----------------------

type IAMRolePolicyAttachmentLister struct{}

func (l *IAMRolePolicyAttachmentLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := iam.New(opts.Session)
	roleParams := &iam.ListRolesInput{}
	resources := make([]resource.Resource, 0)

	for {
		roleResp, err := svc.ListRoles(roleParams)
		if err != nil {
			return nil, err
		}

		for _, listedRole := range roleResp.Roles {
			role, err := GetIAMRole(svc, listedRole.RoleName)
			if err != nil {
				logrus.Errorf("Failed to get listed role %s: %v", *listedRole.RoleName, err)
				continue
			}

			polParams := &iam.ListAttachedRolePoliciesInput{
				RoleName: role.RoleName,
			}

			for {
				polResp, err := svc.ListAttachedRolePolicies(polParams)
				if err != nil {
					logrus.Errorf("failed to list attached policies for role %s: %v",
						*role.RoleName, err)
					break
				}
				for _, pol := range polResp.AttachedPolicies {
					resources = append(resources, &IAMRolePolicyAttachment{
						svc:        svc,
						policyArn:  *pol.PolicyArn,
						policyName: *pol.PolicyName,
						role:       role,
					})
				}

				if !*polResp.IsTruncated {
					break
				}

				polParams.Marker = polResp.Marker
			}
		}

		if !*roleResp.IsTruncated {
			break
		}

		roleParams.Marker = roleResp.Marker
	}

	return resources, nil
}
