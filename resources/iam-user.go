package resources

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	libsettings "github.com/ekristen/libnuke/pkg/settings"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const IAMUserResource = "IAMUser"

func init() {
	registry.Register(&registry.Registration{
		Name:     IAMUserResource,
		Scope:    nuke.Account,
		Resource: &IAMUser{},
		Lister:   &IAMUserLister{},
		Settings: []string{
			"IgnorePermissionBoundary",
		},
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
	svc                    iamiface.IAMAPI
	Name                   *string
	Path                   *string
	UserID                 *string
	CreateDate             *time.Time
	PasswordLastUsed       *time.Time
	Tags                   []*iam.Tag
	HasPermissionBoundary  bool
	PermissionBoundaryARN  *string
	PermissionBoundaryType *string
	settings               *libsettings.Setting
}

func (r *IAMUser) Remove(_ context.Context) error {
	if r.HasPermissionBoundary && r.settings.GetBool("IgnorePermissionBoundary") == false {
		fmt.Println("Removing permission boundary for user", *r.Name)
		_, err := r.svc.DeleteUserPermissionsBoundary(&iam.DeleteUserPermissionsBoundaryInput{
			UserName: r.Name,
		})
		if err != nil {
			return err
		}
	}

	_, err := r.svc.DeleteUser(&iam.DeleteUserInput{
		UserName: r.Name,
	})
	if err != nil {
		return err
	}

	return nil
}

func (r *IAMUser) String() string {
	return *r.Name
}

func (r *IAMUser) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *IAMUser) Settings(settings *libsettings.Setting) {
	r.settings = settings
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
			svc:              svc,
			Name:             user.UserName,
			Path:             user.Path,
			UserID:           user.UserId,
			CreateDate:       user.CreateDate,
			PasswordLastUsed: user.PasswordLastUsed,
			Tags:             user.Tags,
		}

		if user.PermissionsBoundary != nil && user.PermissionsBoundary.PermissionsBoundaryArn != nil {
			resourceUser.HasPermissionBoundary = true
			resourceUser.PermissionBoundaryARN = user.PermissionsBoundary.PermissionsBoundaryArn
			resourceUser.PermissionBoundaryType = user.PermissionsBoundary.PermissionsBoundaryType
		}

		resources = append(resources, resourceUser)
	}

	return resources, nil
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
