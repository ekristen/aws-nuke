package resources

import (
	"context"

	"errors"
	"fmt"
	"strings"

	"github.com/gotidy/ptr"

	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
)

const IAMVirtualMFADeviceResource = "IAMVirtualMFADevice"

func init() {
	registry.Register(&registry.Registration{
		Name:   IAMVirtualMFADeviceResource,
		Scope:  nuke.Account,
		Lister: &IAMVirtualMFADeviceLister{},
	})
}

type IAMVirtualMFADeviceLister struct{}

func (l *IAMVirtualMFADeviceLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := iam.New(opts.Session)

	resp, err := svc.ListVirtualMFADevices(&iam.ListVirtualMFADevicesInput{})
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, 0)
	for _, out := range resp.VirtualMFADevices {
		resources = append(resources, &IAMVirtualMFADevice{
			svc:          svc,
			userID:       out.User.UserId,
			userARN:      out.User.Arn,
			userName:     out.User.UserName,
			serialNumber: out.SerialNumber,
		})
	}

	return resources, nil
}

type IAMVirtualMFADevice struct {
	svc          iamiface.IAMAPI
	userID       *string
	userARN      *string
	userName     *string
	serialNumber *string
}

func (v *IAMVirtualMFADevice) Filter() error {
	isRoot := false
	if ptr.ToString(v.userARN) == fmt.Sprintf("arn:aws:iam::%s:root", ptr.ToString(v.userID)) {
		isRoot = true
	}
	if strings.HasSuffix(ptr.ToString(v.serialNumber), "/root-account-mfa-device") {
		isRoot = true
	}

	if isRoot {
		return errors.New("cannot delete root mfa device")
	}

	return nil
}

func (v *IAMVirtualMFADevice) Remove(_ context.Context) error {
	if _, err := v.svc.DeactivateMFADevice(&iam.DeactivateMFADeviceInput{
		UserName:     v.userName,
		SerialNumber: v.serialNumber,
	}); err != nil {
		return err
	}

	if _, err := v.svc.DeleteVirtualMFADevice(&iam.DeleteVirtualMFADeviceInput{
		SerialNumber: v.serialNumber,
	}); err != nil {
		return err
	}

	return nil
}

func (v *IAMVirtualMFADevice) String() string {
	return ptr.ToString(v.serialNumber)
}
