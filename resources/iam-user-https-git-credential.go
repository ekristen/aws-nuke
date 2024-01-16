package resources

import (
	"context"

	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"

	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const IAMUserHTTPSGitCredentialResource = "IAMUserHTTPSGitCredential"

func init() {
	resource.Register(resource.Registration{
		Name:   IAMUserHTTPSGitCredentialResource,
		Scope:  nuke.Account,
		Lister: &IAMUserHTTPSGitCredentialLister{},
	})
}

type IAMUserHTTPSGitCredential struct {
	svc      iamiface.IAMAPI
	id       string
	userName string
	status   string
	userTags []*iam.Tag
}

func (e *IAMUserHTTPSGitCredential) Remove(_ context.Context) error {
	_, err := e.svc.DeleteServiceSpecificCredential(
		&iam.DeleteServiceSpecificCredentialInput{
			UserName:                    &e.userName,
			ServiceSpecificCredentialId: &e.id,
		})
	if err != nil {
		return err
	}

	return nil
}

func (e *IAMUserHTTPSGitCredential) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("UserName", e.userName)
	properties.Set("ServiceSpecificCredentialId", e.id)

	for _, tag := range e.userTags {
		properties.SetTag(tag.Key, tag.Value)
	}

	return properties
}

func (e *IAMUserHTTPSGitCredential) String() string {
	return fmt.Sprintf("%s -> %s", e.userName, e.id)
}

// --------------

type IAMUserHTTPSGitCredentialLister struct{}

func (l *IAMUserHTTPSGitCredentialLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := iam.New(opts.Session)
	resp, err := svc.ListUsers(nil)
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, 0)
	for _, role := range resp.Users {
		resp, err := svc.ListServiceSpecificCredentials(
			&iam.ListServiceSpecificCredentialsInput{
				UserName:    role.UserName,
				ServiceName: aws.String("codecommit.amazonaws.com"),
			})
		if err != nil {
			return nil, err
		}

		userTags, err := svc.ListUserTags(&iam.ListUserTagsInput{UserName: role.UserName})
		if err != nil {
			return nil, err
		}

		for _, meta := range resp.ServiceSpecificCredentials {
			resources = append(resources, &IAMUserHTTPSGitCredential{
				svc:      svc,
				id:       *meta.ServiceSpecificCredentialId,
				userName: *meta.UserName,
				status:   *meta.Status,
				userTags: userTags.Tags,
			})
		}
	}

	return resources, nil
}
