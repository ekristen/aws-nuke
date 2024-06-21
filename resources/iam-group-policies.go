package resources

import (
	"context"

	"fmt"

	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const IAMGroupPolicyResource = "IAMGroupPolicy"

func init() {
	registry.Register(&registry.Registration{
		Name:   IAMGroupPolicyResource,
		Scope:  nuke.Account,
		Lister: &IAMGroupPolicyLister{},
	})
}

type IAMGroupPolicyLister struct{}

func (l *IAMGroupPolicyLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := iam.New(opts.Session)
	resp, err := svc.ListGroups(nil)
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, 0)
	for _, group := range resp.Groups {
		resp, err := svc.ListGroupPolicies(
			&iam.ListGroupPoliciesInput{
				GroupName: group.GroupName,
			})
		if err != nil {
			return nil, err
		}

		for _, pol := range resp.PolicyNames {
			resources = append(resources, &IAMGroupPolicy{
				svc:        svc,
				policyName: *pol,
				groupName:  *group.GroupName,
			})
		}
	}

	return resources, nil
}

type IAMGroupPolicy struct {
	svc        iamiface.IAMAPI
	policyName string
	groupName  string
}

func (e *IAMGroupPolicy) Remove(_ context.Context) error {
	_, err := e.svc.DeleteGroupPolicy(
		&iam.DeleteGroupPolicyInput{
			PolicyName: &e.policyName,
			GroupName:  &e.groupName,
		})
	if err != nil {
		return err
	}

	return nil
}

func (e *IAMGroupPolicy) String() string {
	return fmt.Sprintf("%s -> %s", e.groupName, e.policyName)
}
