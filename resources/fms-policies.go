package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/fms"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const FMSPolicyResource = "FMSPolicy"

func init() {
	registry.Register(&registry.Registration{
		Name:     FMSPolicyResource,
		Scope:    nuke.Account,
		Resource: &FMSPolicy{},
		Lister:   &FMSPolicyLister{},
	})
}

type FMSPolicyLister struct{}

func (l *FMSPolicyLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := fms.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &fms.ListPoliciesInput{
		MaxResults: aws.Int64(50),
	}

	for {
		resp, err := svc.ListPolicies(params)
		if err != nil {
			return nil, err
		}

		for _, policy := range resp.PolicyList {
			resources = append(resources, &FMSPolicy{
				svc:    svc,
				policy: policy,
			})
		}

		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return resources, nil
}

type FMSPolicy struct {
	svc    *fms.FMS
	policy *fms.PolicySummary
}

func (f *FMSPolicy) Remove(_ context.Context) error {
	_, err := f.svc.DeletePolicy(&fms.DeletePolicyInput{
		PolicyId:                 f.policy.PolicyId,
		DeleteAllPolicyResources: aws.Bool(false),
	})

	return err
}

func (f *FMSPolicy) String() string {
	return *f.policy.PolicyId
}

func (f *FMSPolicy) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("PolicyID", f.policy.PolicyId)
	properties.Set("PolicyName", f.policy.PolicyName)
	return properties
}
