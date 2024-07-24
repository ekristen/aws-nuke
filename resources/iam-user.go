package resources

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const IAMUserResource = "IAMUser"

func init() {
	registry.Register(&registry.Registration{
		Name:   IAMUserResource,
		Scope:  nuke.Account,
		Lister: &IAMUserLister{},
		DependsOn: []string{
			IAMUserAccessKeyResource,
			IAMUserHTTPSGitCredentialResource,
			IAMUserGroupAttachmentResource,
			IAMUserPolicyAttachmentResource,
			IAMVirtualMFADeviceResource,
		},
		DeprecatedAliases: []string{
			"IamUser", // TODO(v4): remove
		},
	})
}

type IAMUser struct {
	svc                   iamiface.IAMAPI
	id                    *string
	name                  *string
	hasPermissionBoundary bool
	createDate            *time.Time
	tags                  []*iam.Tag
}

func (r *IAMUser) Remove(_ context.Context) error {
	if r.hasPermissionBoundary {
		_, err := r.svc.DeleteUserPermissionsBoundary(&iam.DeleteUserPermissionsBoundaryInput{
			UserName: r.name,
		})
		if err != nil {
			return err
		}
	}

	_, err := r.svc.DeleteUser(&iam.DeleteUserInput{
		UserName: r.name,
	})
	if err != nil {
		return err
	}

	return nil
}

func (r *IAMUser) String() string {
	return *r.name
}

func (r *IAMUser) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("UserID", r.id)
	properties.Set("Name", r.name)
	properties.Set("HasPermissionBoundary", r.hasPermissionBoundary)
	properties.Set("CreateDate", r.createDate.Format(time.RFC3339))

	for _, tag := range r.tags {
		properties.SetTag(tag.Key, tag.Value)
	}

	return properties
}

// --------------

// GetIAMUser retries and returns just the *iam.User from the response
func GetIAMUser(svc iamiface.IAMAPI, userName *string) (*iam.User, error) {
	resp, err := svc.GetUser(&iam.GetUserInput{
		UserName: userName,
	})
	if err != nil {
		return nil, err
	}

	return resp.User, err
}

// ListIAMUsers retrieves a base list of users
func ListIAMUsers(svc iamiface.IAMAPI) ([]*iam.User, error) {
	var users []*iam.User
	if err := svc.ListUsersPages(nil, func(page *iam.ListUsersOutput, lastPage bool) bool {
		users = append(users, page.Users...)
		return true
	}); err != nil {
		return nil, err
	}

	return users, nil
}

// --------------

type IAMUserLister struct {
	mockSvc iamiface.IAMAPI
}

func (l *IAMUserLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	var resources []resource.Resource

	var svc iamiface.IAMAPI
	if l.mockSvc != nil {
		svc = l.mockSvc
	} else {
		svc = iam.New(opts.Session)
	}

	allUsers, err := ListIAMUsers(svc)
	if err != nil {
		return resources, err
	}

	for _, out := range allUsers {
		// Note: we have to do a GetIAMUser because the listing of users does not include all the information we need
		user, getErr := GetIAMUser(svc, out.UserName)
		if getErr != nil {
			logrus.Errorf("failed to get user %s: %v", *out.UserName, err)
			continue
		}

		resourceUser := &IAMUser{
			svc:        svc,
			id:         user.UserId,
			name:       user.UserName,
			createDate: user.CreateDate,
			tags:       user.Tags,
		}

		if user.PermissionsBoundary != nil && user.PermissionsBoundary.PermissionsBoundaryArn != nil {
			resourceUser.hasPermissionBoundary = true
		}

		resources = append(resources, resourceUser)
	}

	return resources, nil
}
