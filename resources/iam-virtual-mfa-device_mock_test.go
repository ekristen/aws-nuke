package resources

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"

	"github.com/aws/aws-sdk-go/service/iam" //nolint:staticcheck

	"github.com/ekristen/aws-nuke/v3/mocks/mock_iamiface"
)

func Test_Mock_IAMVirtualMFADevice_List(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockIAM := mock_iamiface.NewMockIAMAPI(ctrl)

	mockIAM.EXPECT().ListVirtualMFADevices(gomock.Any()).Return(&iam.ListVirtualMFADevicesOutput{
		VirtualMFADevices: []*iam.VirtualMFADevice{
			{
				SerialNumber: ptr.String("serial:device1"),
				User: &iam.User{
					UserName: ptr.String("user1"),
					Arn:      ptr.String("arn:aws:iam::123456789012:user/user1"),
				},
			},
			{
				SerialNumber: ptr.String("arn:aws:iam::077097111583:mfa/Authenticator"),
				User: &iam.User{
					UserName: ptr.String("user1"),
					UserId:   ptr.String("0000000000000"),
					Arn:      ptr.String("arn:aws:iam::123456789012:user/user1"),
				},
			},
			{
				SerialNumber: ptr.String("serial:device2"),
				User:         nil,
			},
		},
	}, nil)

	lister := &IAMVirtualMFADeviceLister{
		mockSvc: mockIAM,
	}

	resources, err := lister.List(context.TODO(), testListerOpts)
	a.Nil(err)
	a.Len(resources, 3)
}

func Test_IAMVirtualMFADevice_Properties(t *testing.T) {
	a := assert.New(t)

	iamVirtualMFADevice := IAMVirtualMFADevice{
		user: &iam.User{
			UserName: ptr.String("foobar"),
			Arn:      ptr.String("arn:aws:iam::123456789012:user/foobar"),
		},
		SerialNumber: ptr.String("serial:foobar"),
		Assigned:     ptr.Bool(true),
	}

	properties := iamVirtualMFADevice.Properties()
	a.Equal("serial:foobar", properties.Get("SerialNumber"))
	a.Equal("true", properties.Get("Assigned"))
	a.Equal("serial:foobar", iamVirtualMFADevice.String())
}

func Test_IAMVirtualMFADevice_Filter(t *testing.T) {
	a := assert.New(t)

	rootMFADevice := &IAMVirtualMFADevice{
		user: &iam.User{
			UserId: ptr.String("0000000000000"),
			Arn:    ptr.String("arn:aws:iam::0000000000000:root"),
		},
		SerialNumber: ptr.String("arn:aws:iam::0000000000000:mfa/root-account-mfa-device"),
	}

	err := rootMFADevice.Filter()
	a.NotNil(err)
	a.EqualError(err, "cannot delete root mfa device")

	nonRootMFADevice := &IAMVirtualMFADevice{
		user: &iam.User{
			UserId: ptr.String("123456789012"),
			Arn:    ptr.String("arn:aws:iam::123456789012:user/user1"),
		},
		SerialNumber: ptr.String("arn:aws:iam::123456789012:mfa/user1"),
	}

	err = nonRootMFADevice.Filter()
	a.Nil(err)
}

func Test_Mock_IAMVirtualMFADevice_Remove(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockIAM := mock_iamiface.NewMockIAMAPI(ctrl)

	iamVirtualMFADevice := IAMVirtualMFADevice{
		svc: mockIAM,
		user: &iam.User{
			UserName: ptr.String("foobar"),
			Arn:      ptr.String("arn:aws:iam::123456789012:user/foobar"),
		},
		SerialNumber: ptr.String("serial:foobar"),
	}

	mockIAM.EXPECT().DeactivateMFADevice(gomock.Eq(&iam.DeactivateMFADeviceInput{
		UserName:     iamVirtualMFADevice.user.UserName,
		SerialNumber: iamVirtualMFADevice.SerialNumber,
	})).Return(&iam.DeactivateMFADeviceOutput{}, nil)

	mockIAM.EXPECT().DeleteVirtualMFADevice(gomock.Eq(&iam.DeleteVirtualMFADeviceInput{
		SerialNumber: iamVirtualMFADevice.SerialNumber,
	})).Return(&iam.DeleteVirtualMFADeviceOutput{}, nil)

	err := iamVirtualMFADevice.Remove(context.TODO())
	a.Nil(err)
}
