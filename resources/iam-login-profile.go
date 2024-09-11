package resources

import (
	"context"
	"errors"
	"time"

	"github.com/gotidy/ptr"
	"github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const IAMLoginProfileResource = "IAMLoginProfile"

func init() {
	registry.Register(&registry.Registration{
		Name:   IAMLoginProfileResource,
		Scope:  nuke.Account,
		Lister: &IAMLoginProfileLister{},
	})
}

type IAMLoginProfileLister struct {
	mockSvc iamiface.IAMAPI
}

func (l *IAMLoginProfileLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	resources := make([]resource.Resource, 0)
	var svc iamiface.IAMAPI

	if l.mockSvc != nil {
		svc = l.mockSvc
	} else {
		svc = iam.New(opts.Session)
	}

	resp, err := svc.ListUsers(nil)
	if err != nil {
		return nil, err
	}

	for _, out := range resp.Users {
		lpresp, err := svc.GetLoginProfile(&iam.GetLoginProfileInput{UserName: out.UserName})
		if err != nil {
			var awsError awserr.Error
			if errors.As(err, &awsError) {
				if awsError.Code() == iam.ErrCodeNoSuchEntityException {
					continue
				}
			}

			logrus.Errorf("failed to list login profile for user %s: %v",
				ptr.ToString(out.UserName), err)
			continue
		}

		if lpresp.LoginProfile != nil {
			resources = append(resources, &IAMLoginProfile{
				svc:        svc,
				UserName:   out.UserName,
				CreateDate: out.CreateDate,
			})
		}
	}

	return resources, nil
}

type IAMLoginProfile struct {
	svc        iamiface.IAMAPI
	UserName   *string
	CreateDate *time.Time
}

func (r *IAMLoginProfile) Remove(_ context.Context) error {
	_, err := r.svc.DeleteLoginProfile(&iam.DeleteLoginProfileInput{
		UserName: r.UserName,
	})
	if err != nil {
		return err
	}
	return nil
}

func (r *IAMLoginProfile) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *IAMLoginProfile) String() string {
	return *r.UserName
}
