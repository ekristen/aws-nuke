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

const IAMUserAccessKeyResource = "IAMUserAccessKey"

func init() {
	resource.Register(resource.Registration{
		Name:   IAMUserAccessKeyResource,
		Scope:  nuke.Account,
		Lister: &IAMUserAccessKeyLister{},
	})
}

type IAMUserAccessKey struct {
	svc         iamiface.IAMAPI
	accessKeyId string
	userName    string
	status      string
	userTags    []*iam.Tag
}

func (e *IAMUserAccessKey) Remove(_ context.Context) error {
	_, err := e.svc.DeleteAccessKey(
		&iam.DeleteAccessKeyInput{
			AccessKeyId: &e.accessKeyId,
			UserName:    &e.userName,
		})
	if err != nil {
		return err
	}

	return nil
}

func (e *IAMUserAccessKey) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("UserName", e.userName)
	properties.Set("AccessKeyID", e.accessKeyId)

	for _, tag := range e.userTags {
		properties.SetTag(tag.Key, tag.Value)
	}

	return properties
}

func (e *IAMUserAccessKey) String() string {
	return fmt.Sprintf("%s -> %s", e.userName, e.accessKeyId)
}

// --------------

type IAMUserAccessKeyLister struct{}

func (l *IAMUserAccessKeyLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := iam.New(opts.Session)
	resp, err := svc.ListUsers(nil)
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, 0)
	for _, role := range resp.Users {
		resp, err := svc.ListAccessKeys(
			&iam.ListAccessKeysInput{
				UserName: role.UserName,
			})
		if err != nil {
			return nil, err
		}

		userTags, err := svc.ListUserTags(&iam.ListUserTagsInput{UserName: role.UserName})
		if err != nil {
			return nil, err
		}

		for _, meta := range resp.AccessKeyMetadata {
			resources = append(resources, &IAMUserAccessKey{
				svc:         svc,
				accessKeyId: *meta.AccessKeyId,
				userName:    *meta.UserName,
				status:      *meta.Status,
				userTags:    userTags.Tags,
			})
		}
	}

	return resources, nil
}
