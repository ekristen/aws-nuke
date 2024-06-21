package resources

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const IAMPolicyResource = "IAMPolicy"

func init() {
	registry.Register(&registry.Registration{
		Name:   IAMPolicyResource,
		Scope:  nuke.Account,
		Lister: &IAMPolicyLister{},
		DependsOn: []string{
			IAMUserPolicyAttachmentResource,
			IAMGroupPolicyAttachmentResource,
			IAMRolePolicyAttachmentResource,
		},
		DeprecatedAliases: []string{
			"IamPolicy",
		},
	})
}

type IAMPolicy struct {
	svc      iamiface.IAMAPI
	name     string
	policyID string
	arn      string
	path     string
	tags     []*iam.Tag
}

func (e *IAMPolicy) Remove(_ context.Context) error {
	resp, err := e.svc.ListPolicyVersions(&iam.ListPolicyVersionsInput{
		PolicyArn: &e.arn,
	})
	if err != nil {
		return err
	}

	for _, version := range resp.Versions {
		if !*version.IsDefaultVersion {
			_, err = e.svc.DeletePolicyVersion(&iam.DeletePolicyVersionInput{
				PolicyArn: &e.arn,
				VersionId: version.VersionId,
			})
			if err != nil {
				return err
			}
		}
	}

	_, err = e.svc.DeletePolicy(&iam.DeletePolicyInput{
		PolicyArn: &e.arn,
	})
	if err != nil {
		return err
	}

	return nil
}

func (e *IAMPolicy) Properties() types.Properties {
	properties := types.NewProperties()

	properties.Set("Name", e.name)
	properties.Set("ARN", e.arn)
	properties.Set("Path", e.path)
	properties.Set("PolicyID", e.policyID)
	for _, tag := range e.tags {
		properties.SetTag(tag.Key, tag.Value)
	}
	return properties
}

func (e *IAMPolicy) String() string {
	return e.arn
}

// -------------

func GetIAMPolicy(svc *iam.IAM, policyArn *string) (*iam.Policy, error) {
	params := &iam.GetPolicyInput{
		PolicyArn: policyArn,
	}
	resp, err := svc.GetPolicy(params)
	return resp.Policy, err
}

type IAMPolicyLister struct{}

func (l *IAMPolicyLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := iam.New(opts.Session)

	params := &iam.ListPoliciesInput{
		Scope: aws.String("Local"),
	}

	policies := make([]*iam.Policy, 0)

	if err := svc.ListPoliciesPages(params,
		func(page *iam.ListPoliciesOutput, lastPage bool) bool {
			for _, listedPolicy := range page.Policies {
				policy, err := GetIAMPolicy(svc, listedPolicy.Arn)
				if err != nil {
					logrus.Errorf("Failed to get listed policy %s: %v", *listedPolicy.PolicyName, err)
					continue
				}
				policies = append(policies, policy)
			}
			return true
		}); err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, 0)

	for _, out := range policies {
		resources = append(resources, &IAMPolicy{
			svc:      svc,
			name:     *out.PolicyName,
			path:     *out.Path,
			arn:      *out.Arn,
			policyID: *out.PolicyId,
			tags:     out.Tags,
		})
	}

	return resources, nil
}
