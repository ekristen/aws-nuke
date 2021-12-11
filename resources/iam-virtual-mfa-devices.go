package resources

import (
	"errors"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"
)

type IAMVirtualMFADevice struct {
	svc          iamiface.IAMAPI
	userName     string
	serialNumber string
}

func init() {
	register("IAMVirtualMFADevice", ListIAMVirtualMFADevices)
}

func ListIAMVirtualMFADevices(sess *session.Session) ([]Resource, error) {
	svc := iam.New(sess)

	resp, err := svc.ListVirtualMFADevices(&iam.ListVirtualMFADevicesInput{})
	if err != nil {
		return nil, err
	}

	resources := make([]Resource, 0)
	for _, out := range resp.VirtualMFADevices {
		resources = append(resources, &IAMVirtualMFADevice{
			svc:          svc,
			userName:     *out.User.UserName,
			serialNumber: *out.SerialNumber,
		})
	}

	return resources, nil
}

func (v *IAMVirtualMFADevice) Filter() error {
	if strings.HasSuffix(v.serialNumber, "/root-account-mfa-device") {
		return errors.New("cannot delete root mfa device")
	}
	return nil
}

func (v *IAMVirtualMFADevice) Remove() error {
	if _, err := v.svc.DeactivateMFADevice(&iam.DeactivateMFADeviceInput{
		UserName:     aws.String(v.userName),
		SerialNumber: aws.String(v.serialNumber),
	}); err != nil {
		return err
	}

	if _, err := v.svc.DeleteVirtualMFADevice(&iam.DeleteVirtualMFADeviceInput{
		SerialNumber: &v.serialNumber,
	}); err != nil {
		return err
	}

	return nil
}

func (v *IAMVirtualMFADevice) String() string {
	return v.serialNumber
}
