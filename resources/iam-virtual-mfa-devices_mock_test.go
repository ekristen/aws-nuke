package resources

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/golang/mock/gomock"
	"github.com/rebuy-de/aws-nuke/v2/mocks/mock_iamiface"
	"github.com/stretchr/testify/assert"
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

	err := iamVirtualMFADevice.Remove()
	a.Nil(err)
}
