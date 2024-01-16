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

const IAMUserResource = "IAMUser"

func init() {
	resource.Register(resource.Registration{
		Name:   IAMUserResource,
		Scope:  nuke.Account,
		Lister: &IAMUsersLister{},
		DependsOn: []string{
			IAMUserAccessKeyResource,
			IAMUserHTTPSGitCredentialResource,
			IAMUserGroupAttachmentResource,
			IAMUserPolicyAttachmentResource,
			IAMVirtualMFADeviceResource,
		},
	})
}

type IAMUser struct {
	svc  iamiface.IAMAPI
	name string
	tags []*iam.Tag
}

func (e *IAMUser) Remove(_ context.Context) error {
	_, err := e.svc.DeleteUser(&iam.DeleteUserInput{
		UserName: &e.name,
	})
	if err != nil {
		return err
	}

	return nil
}

func (e *IAMUser) String() string {
	return e.name
}

func (e *IAMUser) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("Name", e.name)

	for _, tag := range e.tags {
		properties.SetTag(tag.Key, tag.Value)
	}

	return properties
}

// --------------

func GetIAMUser(svc *iam.IAM, userName *string) (*iam.User, error) {
	params := &iam.GetUserInput{
		UserName: userName,
	}
	resp, err := svc.GetUser(params)
	return resp.User, err
}

// --------------

type IAMUsersLister struct{}

func (l *IAMUsersLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := iam.New(opts.Session)

	var resources []resource.Resource

	if err := svc.ListUsersPages(nil, func(page *iam.ListUsersOutput, lastPage bool) bool {
		for _, out := range page.Users {
			user, err := GetIAMUser(svc, out.UserName)
			if err != nil {
				logrus.Errorf("Failed to get user %s: %v", *out.UserName, err)
				continue
			}
			resources = append(resources, &IAMUser{
				svc:  svc,
				name: *out.UserName,
				tags: user.Tags,
			})
		}
		return true
	}); err != nil {
		return nil, err
	}

	return resources, nil
}
