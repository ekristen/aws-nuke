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

func (r *IAMVirtualMFADevice) Filter() error {
	isRoot := false
	if ptr.ToString(r.userARN) == fmt.Sprintf("arn:aws:iam::%s:root", ptr.ToString(r.userID)) {
		isRoot = true
	}
	if strings.HasSuffix(ptr.ToString(r.serialNumber), "/root-account-mfa-device") {
		isRoot = true
	}

	if isRoot {
		return errors.New("cannot delete root mfa device")
	}

	return nil
}

func (r *IAMVirtualMFADevice) Remove(_ context.Context) error {
	if _, err := r.svc.DeactivateMFADevice(&iam.DeactivateMFADeviceInput{
		UserName:     r.userName,
		SerialNumber: r.serialNumber,
	}); err != nil {
		return err
	}

	if _, err := r.svc.DeleteVirtualMFADevice(&iam.DeleteVirtualMFADeviceInput{
		SerialNumber: r.serialNumber,
	}); err != nil {
		return err
	}

	return nil
}

func (r *IAMVirtualMFADevice) String() string {
	return ptr.ToString(r.serialNumber)
}
