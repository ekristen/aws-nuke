package resources

import (
	"context"

	"fmt"

	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"

	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const IAMUserPolicyResource = "IAMUserPolicy"

func init() {
	resource.Register(&resource.Registration{
		Name:   IAMUserPolicyResource,
		Scope:  nuke.Account,
		Lister: &IAMUserPolicyLister{},
	})
}

type IAMUserPolicy struct {
	svc        iamiface.IAMAPI
	userName   string
	policyName string
}

func (e *IAMUserPolicy) Remove(_ context.Context) error {
	_, err := e.svc.DeleteUserPolicy(
		&iam.DeleteUserPolicyInput{
			UserName:   &e.userName,
			PolicyName: &e.policyName,
		})
	if err != nil {
		return err
	}

	return nil
}

func (e *IAMUserPolicy) String() string {
	return fmt.Sprintf("%s -> %s", e.userName, e.policyName)
}

// ----------------

type IAMUserPolicyLister struct{}

func (l *IAMUserPolicyLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := iam.New(opts.Session)

	users, err := svc.ListUsers(nil)
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, 0)
	for _, user := range users.Users {
		policies, err := svc.ListUserPolicies(&iam.ListUserPoliciesInput{
			UserName: user.UserName,
		})
		if err != nil {
			return nil, err
		}

		for _, policyName := range policies.PolicyNames {
			resources = append(resources, &IAMUserPolicy{
				svc:        svc,
				policyName: *policyName,
				userName:   *user.UserName,
			})
		}
	}

	return resources, nil
}
