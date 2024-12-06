package resources

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/gotidy/ptr"
	"github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const IAMVirtualMFADeviceResource = "IAMVirtualMFADevice"

func init() {
	registry.Register(&registry.Registration{
		Name:     IAMVirtualMFADeviceResource,
		Scope:    nuke.Account,
		Resource: &IAMVirtualMFADevice{},
		Lister:   &IAMVirtualMFADeviceLister{},
	})
}

type IAMVirtualMFADeviceLister struct {
	mockSvc iamiface.IAMAPI
}

func (l *IAMVirtualMFADeviceLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	resources := make([]resource.Resource, 0)

	var svc iamiface.IAMAPI
	if l.mockSvc != nil {
		svc = l.mockSvc
	} else {
		svc = iam.New(opts.Session)
	}

	resp, err := svc.ListVirtualMFADevices(&iam.ListVirtualMFADevicesInput{})
	if err != nil {
		return nil, err
	}

	for _, out := range resp.VirtualMFADevices {
		resources = append(resources, &IAMVirtualMFADevice{
			svc:          svc,
			user:         out.User,
			SerialNumber: out.SerialNumber,
			Assigned:     ptr.Bool(out.User != nil),
		})
	}

	return resources, nil
}

type IAMVirtualMFADevice struct {
	svc          iamiface.IAMAPI
	user         *iam.User
	Assigned     *bool
	SerialNumber *string
}

func (r *IAMVirtualMFADevice) Filter() error {
	isRoot := false
	if r.user != nil && ptr.ToString(r.user.Arn) == fmt.Sprintf("arn:aws:iam::%s:root", ptr.ToString(r.user.UserId)) {
		logrus.Debug("user is not nil, arn is root, assuming root")
		isRoot = true
	}
	if !isRoot && strings.HasSuffix(ptr.ToString(r.SerialNumber), "/root-account-mfa-device") {
		logrus.Debug("serial number is root, assuming root")
		isRoot = true
	}

	if isRoot {
		return errors.New("cannot delete root mfa device")
	}

	return nil
}

func (r *IAMVirtualMFADevice) Remove(_ context.Context) error {
	// Note: if the user is not nil, we need to deactivate the MFA device first
	if r.user != nil {
		if _, err := r.svc.DeactivateMFADevice(&iam.DeactivateMFADeviceInput{
			UserName:     r.user.UserName,
			SerialNumber: r.SerialNumber,
		}); err != nil {
			return err
		}
	}

	if _, err := r.svc.DeleteVirtualMFADevice(&iam.DeleteVirtualMFADeviceInput{
		SerialNumber: r.SerialNumber,
	}); err != nil {
		return err
	}

	return nil
}

func (r *IAMVirtualMFADevice) String() string {
	return ptr.ToString(r.SerialNumber)
}

func (r *IAMVirtualMFADevice) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}
