package resources

import (
	"context"

	"errors"

	"github.com/gotidy/ptr"
	"github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"

	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const IAMLoginProfileResource = "IAMLoginProfile"

func init() {
	resource.Register(&resource.Registration{
		Name:   IAMLoginProfileResource,
		Scope:  nuke.Account,
		Lister: &IAMLoginProfileLister{},
	})
}

type IAMLoginProfileLister struct{}

func (l *IAMLoginProfileLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := iam.New(opts.Session)

	resp, err := svc.ListUsers(nil)
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, 0)
	for _, out := range resp.Users {
		lpresp, err := svc.GetLoginProfile(&iam.GetLoginProfileInput{UserName: out.UserName})
		if err != nil {
			var awsError awserr.Error
			if errors.As(err, &awsError) {
				switch awsError.Code() {
				case iam.ErrCodeNoSuchEntityException:
					// The user does not have a login profile and we do not
					// need to print an error for that.
					continue
				}
			}

			logrus.Errorf("failed to list login profile for user %s: %v",
				ptr.ToString(out.UserName), err)
			continue
		}

		if lpresp.LoginProfile != nil {
			resources = append(resources, &IAMLoginProfile{
				svc:  svc,
				name: ptr.ToString(out.UserName),
			})
		}
	}

	return resources, nil
}

type IAMLoginProfile struct {
	svc  iamiface.IAMAPI
	name string
}

func (e *IAMLoginProfile) Remove(_ context.Context) error {
	_, err := e.svc.DeleteLoginProfile(&iam.DeleteLoginProfileInput{UserName: &e.name})
	if err != nil {
		return err
	}
	return nil
}

func (e *IAMLoginProfile) Properties() types.Properties {
	return types.NewProperties().
		Set("UserName", e.name)
}

func (e *IAMLoginProfile) String() string {
	return e.name
}
