package resources

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws"         //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/iam" //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/iam/iamiface"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const IAMPolicyResource = "IAMPolicy"

func init() {
	registry.Register(&registry.Registration{
		Name:     IAMPolicyResource,
		Scope:    nuke.Account,
		Resource: &IAMPolicy{},
		Lister:   &IAMPolicyLister{},
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
			svc:        svc,
			Name:       out.PolicyName,
			Path:       out.Path,
			ARN:        out.Arn,
			PolicyID:   out.PolicyId,
			CreateDate: out.CreateDate,
			Tags:       out.Tags,
		})
	}

	return resources, nil
}

type IAMPolicy struct {
	svc        iamiface.IAMAPI
	Name       *string
	PolicyID   *string
	ARN        *string
	Path       *string
	CreateDate *time.Time
	Tags       []*iam.Tag
}

func (r *IAMPolicy) Remove(_ context.Context) error {
	resp, err := r.svc.ListPolicyVersions(&iam.ListPolicyVersionsInput{
		PolicyArn: r.ARN,
	})
	if err != nil {
		return err
	}

	for _, version := range resp.Versions {
		if !*version.IsDefaultVersion {
			_, err = r.svc.DeletePolicyVersion(&iam.DeletePolicyVersionInput{
				PolicyArn: r.ARN,
				VersionId: version.VersionId,
			})
			if err != nil {
				return err
			}
		}
	}

	_, err = r.svc.DeletePolicy(&iam.DeletePolicyInput{
		PolicyArn: r.ARN,
	})
	if err != nil {
		return err
	}

	return nil
}

func (r *IAMPolicy) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *IAMPolicy) String() string {
	return *r.ARN
}

// -------------

func GetIAMPolicy(svc *iam.IAM, policyArn *string) (*iam.Policy, error) {
	resp, err := svc.GetPolicy(&iam.GetPolicyInput{
		PolicyArn: policyArn,
	})
	if err != nil {
		return nil, err
	}

	return resp.Policy, nil
}
