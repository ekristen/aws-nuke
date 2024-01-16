package resources

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"

	"github.com/ekristen/aws-nuke/mocks/mock_iamiface"
)

func Test_Mock_IAMVirtualMFADevice_Remove(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockIAM := mock_iamiface.NewMockIAMAPI(ctrl)

	iamVirtualMFADevice := IAMVirtualMFADevice{
		svc:          mockIAM,
		userName:     "user:foobar",
		serialNumber: "serial:foobar",
	}

	mockIAM.EXPECT().DeactivateMFADevice(gomock.Eq(&iam.DeactivateMFADeviceInput{
		UserName:     &iamVirtualMFADevice.userName,
		SerialNumber: &iamVirtualMFADevice.serialNumber,
	})).Return(&iam.DeactivateMFADeviceOutput{}, nil)

	mockIAM.EXPECT().DeleteVirtualMFADevice(gomock.Eq(&iam.DeleteVirtualMFADeviceInput{
		SerialNumber: aws.String(iamVirtualMFADevice.serialNumber),
	})).Return(&iam.DeleteVirtualMFADeviceOutput{}, nil)

	err := iamVirtualMFADevice.Remove(context.TODO())
	a.Nil(err)
}
