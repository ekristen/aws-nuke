package resources

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const IAMAccountSettingPasswordPolicyResource = "IAMAccountSettingPasswordPolicy"

func init() {
	registry.Register(&registry.Registration{
		Name:     IAMAccountSettingPasswordPolicyResource,
		Scope:    nuke.Account,
		Resource: &IAMAccountSettingPasswordPolicy{},
		Lister:   &IAMAccountSettingPasswordPolicyLister{},
	})
}

type IAMAccountSettingPasswordPolicyLister struct{}

func (l *IAMAccountSettingPasswordPolicyLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	resources := make([]resource.Resource, 0)

	svc := iam.New(opts.Session)

	resp, err := svc.GetAccountPasswordPolicy(&iam.GetAccountPasswordPolicyInput{})

	if err != nil {
		var aerr awserr.Error
		if errors.As(err, &aerr) {
			switch aerr.Code() {
			case iam.ErrCodeNoSuchEntityException:
				return nil, nil
			case iam.ErrCodeServiceFailureException:
				return nil, aerr
			}
		}

		return nil, err
	}

	resources = append(resources, &IAMAccountSettingPasswordPolicy{
		svc:    svc,
		policy: resp.PasswordPolicy,
	})

	return resources, nil
}

type IAMAccountSettingPasswordPolicy struct {
	svc    iamiface.IAMAPI
	policy *iam.PasswordPolicy
}

func (r *IAMAccountSettingPasswordPolicy) Remove(_ context.Context) error {
	_, err := r.svc.DeleteAccountPasswordPolicy(&iam.DeleteAccountPasswordPolicyInput{})
	if err != nil {
		return err
	}

	return nil
}

func (r *IAMAccountSettingPasswordPolicy) String() string {
	return "custom"
}

func (r *IAMAccountSettingPasswordPolicy) Properties() types.Properties {
	return types.NewProperties().Set("type", "custom")
}
