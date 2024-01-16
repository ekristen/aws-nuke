package resources

import (
	"context"

	"fmt"

	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"

	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const IAMUserGroupAttachmentResource = "IAMUserGroupAttachment"

func init() {
	resource.Register(resource.Registration{
		Name:   IAMUserGroupAttachmentResource,
		Scope:  nuke.Account,
		Lister: &IAMUserGroupAttachmentLister{},
	})
}

type IAMUserGroupAttachment struct {
	svc       iamiface.IAMAPI
	groupName string
	userName  string
}

func (e *IAMUserGroupAttachment) Remove(_ context.Context) error {
	_, err := e.svc.RemoveUserFromGroup(
		&iam.RemoveUserFromGroupInput{
			GroupName: &e.groupName,
			UserName:  &e.userName,
		})
	if err != nil {
		return err
	}

	return nil
}

func (e *IAMUserGroupAttachment) String() string {
	return fmt.Sprintf("%s -> %s", e.userName, e.groupName)
}

func (e *IAMUserGroupAttachment) Properties() types.Properties {
	return types.NewProperties().
		Set("GroupName", e.groupName).
		Set("UserName", e.userName)
}

// ------------------------------

type IAMUserGroupAttachmentLister struct{}

func (l *IAMUserGroupAttachmentLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := iam.New(opts.Session)
	resp, err := svc.ListUsers(nil)
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, 0)
	for _, role := range resp.Users {
		resp, err := svc.ListGroupsForUser(
			&iam.ListGroupsForUserInput{
				UserName: role.UserName,
			})
		if err != nil {
			return nil, err
		}

		for _, grp := range resp.Groups {
			resources = append(resources, &IAMUserGroupAttachment{
				svc:       svc,
				groupName: *grp.GroupName,
				userName:  *role.UserName,
			})
		}
	}

	return resources, nil
}
