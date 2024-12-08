package resources

import (
	"context"

	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const IAMUserPolicyAttachmentResource = "IAMUserPolicyAttachment"

func init() {
	registry.Register(&registry.Registration{
		Name:     IAMUserPolicyAttachmentResource,
		Scope:    nuke.Account,
		Resource: &IAMUserPolicyAttachment{},
		Lister:   &IAMUserPolicyAttachmentLister{},
		DeprecatedAliases: []string{
			"IamUserPolicyAttachement",
		},
	})
}

type IAMUserPolicyAttachment struct {
	svc        iamiface.IAMAPI
	policyArn  string
	policyName string
	userName   string
	userTags   []*iam.Tag
}

func (e *IAMUserPolicyAttachment) Remove(_ context.Context) error {
	_, err := e.svc.DetachUserPolicy(
		&iam.DetachUserPolicyInput{
			PolicyArn: &e.policyArn,
			UserName:  &e.userName,
		})
	if err != nil {
		return err
	}

	return nil
}

func (e *IAMUserPolicyAttachment) Properties() types.Properties {
	properties := types.NewProperties().
		Set("PolicyArn", e.policyArn).
		Set("PolicyName", e.policyName).
		Set("UserName", e.userName)
	for _, tag := range e.userTags {
		properties.SetTagWithPrefix("user", tag.Key, tag.Value)
	}
	return properties
}

func (e *IAMUserPolicyAttachment) String() string {
	return fmt.Sprintf("%s -> %s", e.userName, e.policyName)
}

// -------------------------------

type IAMUserPolicyAttachmentLister struct{}

func (l *IAMUserPolicyAttachmentLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := iam.New(opts.Session)

	resp, err := svc.ListUsers(nil)
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, 0)
	for _, user := range resp.Users {
		iamUser, err := GetIAMUser(svc, user.UserName)
		if err != nil {
			logrus.Errorf("Failed to get user %s: %v", *user.UserName, err)
			continue
		}

		resp, err := svc.ListAttachedUserPolicies(
			&iam.ListAttachedUserPoliciesInput{
				UserName: user.UserName,
			})
		if err != nil {
			return nil, err
		}

		for _, pol := range resp.AttachedPolicies {
			resources = append(resources, &IAMUserPolicyAttachment{
				svc:        svc,
				policyArn:  *pol.PolicyArn,
				policyName: *pol.PolicyName,
				userName:   *user.UserName,
				userTags:   iamUser.Tags,
			})
		}
	}

	return resources, nil
}
