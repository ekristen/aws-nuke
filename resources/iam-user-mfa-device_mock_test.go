package resources

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"

	"github.com/aws/aws-sdk-go/service/iam"

	"github.com/ekristen/aws-nuke/v3/mocks/mock_iamiface"
)

func Test_Mock_IAMUserMFADevice_List(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockIAM := mock_iamiface.NewMockIAMAPI(ctrl)

	mockIAM.EXPECT().ListUsersPages(gomock.Any(), gomock.Any()).
		Do(func(_ *iam.ListUsersInput, fn func(page *iam.ListUsersOutput, lastPage bool) bool) {
			fn(&iam.ListUsersOutput{
				Users: []*iam.User{
					{
						UserName: ptr.String("foo"),
					},
				},
			}, true)
		}).Return(nil)

	mockIAM.EXPECT().ListMFADevices(&iam.ListMFADevicesInput{
		UserName: ptr.String("foo"),
	}).Return(&iam.ListMFADevicesOutput{
		MFADevices: []*iam.MFADevice{
			{
				UserName:     ptr.String("foo"),
				SerialNumber: ptr.String("bar"),
				EnableDate:   ptr.Time(time.Now()),
			},
		},
	}, nil)

	lister := IAMUserMFADeviceLister{
		mockSvc: mockIAM,
	}

	resources, err := lister.List(context.TODO(), testListerOpts)
	a.NoError(err)
	a.Len(resources, 1)
}

func Test_Mock_IAMUserMFADevice_Remove(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockIAM := mock_iamiface.NewMockIAMAPI(ctrl)

	mockIAM.EXPECT().DeactivateMFADevice(&iam.DeactivateMFADeviceInput{
		UserName:     ptr.String("foo"),
		SerialNumber: ptr.String("bar"),
	})

	resource := &IAMUserMFADevice{
		svc:          mockIAM,
		UserName:     ptr.String("foo"),
		SerialNumber: ptr.String("bar"),
	}

	err := resource.Remove(context.TODO())
	a.NoError(err)
}
