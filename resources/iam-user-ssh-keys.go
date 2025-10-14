package resources

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/service/iam" //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/iam/iamiface"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const IAMUserSSHPublicKeyResource = "IAMUserSSHPublicKey"

func init() {
	registry.Register(&registry.Registration{
		Name:     IAMUserSSHPublicKeyResource,
		Scope:    nuke.Account,
		Resource: &IAMUserSSHKey{},
		Lister:   &IAMUserSSHPublicKeyLister{},
	})
}

type IAMUserSSHPublicKeyLister struct{}

func (l *IAMUserSSHPublicKeyLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := iam.New(opts.Session)

	usersOutput, err := svc.ListUsers(nil)
	if err != nil {
		return nil, err
	}

	var resources []resource.Resource
	for _, user := range usersOutput.Users {
		listOutput, err := svc.ListSSHPublicKeys(&iam.ListSSHPublicKeysInput{
			UserName: user.UserName,
		})

		if err != nil {
			return nil, err
		}

		for _, publicKey := range listOutput.SSHPublicKeys {
			resources = append(resources, &IAMUserSSHKey{
				svc:      svc,
				userName: *user.UserName,
				sshKeyID: *publicKey.SSHPublicKeyId,
			})
		}
	}

	return resources, nil
}

type IAMUserSSHKey struct {
	svc      iamiface.IAMAPI
	userName string
	sshKeyID string
}

func (u *IAMUserSSHKey) Properties() types.Properties {
	return types.NewProperties().
		Set("UserName", u.userName).
		Set("SSHKeyID", u.sshKeyID)
}

func (u *IAMUserSSHKey) String() string {
	return fmt.Sprintf("%s -> %s", u.userName, u.sshKeyID)
}

func (u *IAMUserSSHKey) Remove(_ context.Context) error {
	_, err := u.svc.DeleteSSHPublicKey(&iam.DeleteSSHPublicKeyInput{
		UserName:       &u.userName,
		SSHPublicKeyId: &u.sshKeyID,
	})

	return err
}
