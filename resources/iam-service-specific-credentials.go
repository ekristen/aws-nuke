package resources

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"

	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const IAMServiceSpecificCredentialResource = "IAMServiceSpecificCredential"

func init() {
	resource.Register(resource.Registration{
		Name:   IAMServiceSpecificCredentialResource,
		Scope:  nuke.Account,
		Lister: &IAMServiceSpecificCredentialLister{},
	})
}

type IAMServiceSpecificCredentialLister struct{}

func (l *IAMServiceSpecificCredentialLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := iam.New(opts.Session)

	userLister := &IAMUsersLister{}
	users, usersErr := userLister.List(ctx, o)
	if usersErr != nil {
		return nil, usersErr
	}

	resources := make([]resource.Resource, 0)
	for _, userResource := range users {
		user, ok := userResource.(*IAMUser)
		if !ok {
			logrus.Errorf("Unable to cast IAMUser.")
			continue
		}
		params := &iam.ListServiceSpecificCredentialsInput{
			UserName: &user.name,
		}
		serviceCredentials, err := svc.ListServiceSpecificCredentials(params)
		if err != nil {
			return nil, err
		}

		for _, credential := range serviceCredentials.ServiceSpecificCredentials {
			resources = append(resources, &IAMServiceSpecificCredential{
				svc:         svc,
				name:        *credential.ServiceUserName,
				serviceName: *credential.ServiceName,
				id:          *credential.ServiceSpecificCredentialId,
				userName:    user.name,
			})
		}
	}

	return resources, nil
}

type IAMServiceSpecificCredential struct {
	svc         iamiface.IAMAPI
	name        string
	serviceName string
	id          string
	userName    string
}

func (e *IAMServiceSpecificCredential) Remove(_ context.Context) error {
	params := &iam.DeleteServiceSpecificCredentialInput{
		ServiceSpecificCredentialId: &e.id,
		UserName:                    &e.userName,
	}
	_, err := e.svc.DeleteServiceSpecificCredential(params)
	if err != nil {
		return err
	}
	return nil
}

func (e *IAMServiceSpecificCredential) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("ServiceName", e.serviceName)
	properties.Set("ID", e.id)
	return properties
}

func (e *IAMServiceSpecificCredential) String() string {
	return e.userName + " -> " + e.name
}
