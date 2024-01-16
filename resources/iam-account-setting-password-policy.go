package resources

import (
	"context"

	"errors"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"

	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const IAMAccountSettingPasswordPolicyResource = "IAMAccountSettingPasswordPolicy"

func init() {
	resource.Register(resource.Registration{
		Name:   IAMAccountSettingPasswordPolicyResource,
		Scope:  nuke.Account,
		Lister: &IAMAccountSettingPasswordPolicyLister{},
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

func (e *IAMAccountSettingPasswordPolicy) Remove(_ context.Context) error {
	_, err := e.svc.DeleteAccountPasswordPolicy(&iam.DeleteAccountPasswordPolicyInput{})
	if err != nil {
		return err
	}

	return nil
}

func (e *IAMAccountSettingPasswordPolicy) String() string {
	return "custom"
}
