package resources

import (
	"context"

	"fmt"

	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"

	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const IAMGroupPolicyAttachmentResource = "IAMGroupPolicyAttachment"

func init() {
	resource.Register(&resource.Registration{
		Name:   IAMGroupPolicyAttachmentResource,
		Scope:  nuke.Account,
		Lister: &IAMGroupPolicyAttachmentLister{},
		DeprecatedAliases: []string{
			"IamGroupPolicyAttachement",
		},
	})
}

type IAMGroupPolicyAttachmentLister struct{}

func (l *IAMGroupPolicyAttachmentLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := iam.New(opts.Session)

	resp, err := svc.ListGroups(nil)
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, 0)
	for _, role := range resp.Groups {
		resp, err := svc.ListAttachedGroupPolicies(
			&iam.ListAttachedGroupPoliciesInput{
				GroupName: role.GroupName,
			})
		if err != nil {
			return nil, err
		}

		for _, pol := range resp.AttachedPolicies {
			resources = append(resources, &IAMGroupPolicyAttachment{
				svc:        svc,
				policyArn:  *pol.PolicyArn,
				policyName: *pol.PolicyName,
				groupName:  *role.GroupName,
			})
		}
	}

	return resources, nil
}

type IAMGroupPolicyAttachment struct {
	svc        iamiface.IAMAPI
	policyArn  string
	policyName string
	groupName  string
}

func (e *IAMGroupPolicyAttachment) Remove(_ context.Context) error {
	_, err := e.svc.DetachGroupPolicy(
		&iam.DetachGroupPolicyInput{
			PolicyArn: &e.policyArn,
			GroupName: &e.groupName,
		})
	if err != nil {
		return err
	}

	return nil
}

func (e *IAMGroupPolicyAttachment) Properties() types.Properties {
	return types.NewProperties().
		Set("GroupName", e.groupName).
		Set("PolicyName", e.policyName).
		Set("PolicyArn", e.policyArn)
}

func (e *IAMGroupPolicyAttachment) String() string {
	return fmt.Sprintf("%s -> %s", e.groupName, e.policyName)
}
