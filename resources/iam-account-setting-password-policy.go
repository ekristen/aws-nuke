package resources

import (
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"
)

type IAMAccountSettingPasswordPolicy struct {
	svc    iamiface.IAMAPI
	policy *iam.PasswordPolicy
}

func init() {
	register("IAMAccountSettingPasswordPolicy", ListIAMAccountSettingPasswordPolicy)
}

func ListIAMAccountSettingPasswordPolicy(sess *session.Session) ([]Resource, error) {
	resources := make([]Resource, 0)

	svc := iam.New(sess)

	resp, err := svc.GetAccountPasswordPolicy(&iam.GetAccountPasswordPolicyInput{})

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case iam.ErrCodeNoSuchEntityException:
				return nil, nil
			case iam.ErrCodeServiceFailureException:
				return nil, aerr
			default:
				return nil, aerr
			}
		}
	}

	resources = append(resources, &IAMAccountSettingPasswordPolicy{
		svc:    svc,
		policy: resp.PasswordPolicy,
	})

	return resources, nil
}

func (e *IAMAccountSettingPasswordPolicy) Remove() error {
	_, err := e.svc.DeleteAccountPasswordPolicy(&iam.DeleteAccountPasswordPolicyInput{})
	if err != nil {
		return err
	}
	return nil
}

func (e *IAMAccountSettingPasswordPolicy) String() string {
	return "custom"
}
