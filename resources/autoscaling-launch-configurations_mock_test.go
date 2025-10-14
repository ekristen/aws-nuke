package resources

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"

	"github.com/aws/aws-sdk-go/service/autoscaling" //nolint:staticcheck

	"github.com/ekristen/aws-nuke/v3/mocks/mock_autoscalingiface"
)

func Test_Mock_AutoScalingLaunchConfiguration_List(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mock_autoscalingiface.NewMockAutoScalingAPI(ctrl)

	mockSvc.EXPECT().
		DescribeLaunchConfigurationsPages(gomock.Eq(&autoscaling.DescribeLaunchConfigurationsInput{}), gomock.Any()).
		Do(func(input *autoscaling.DescribeLaunchConfigurationsInput,
			fn func(res *autoscaling.DescribeLaunchConfigurationsOutput, lastPage bool) bool) {
			fn(&autoscaling.DescribeLaunchConfigurationsOutput{
				LaunchConfigurations: []*autoscaling.LaunchConfiguration{
					{
						LaunchConfigurationName: ptr.String("foo"),
						CreatedTime:             ptr.Time(time.Now()),
					},
				},
			}, true)
		}).
		Return(nil)

	lister := AutoScalingLaunchConfigurationLister{
		mockSvc: mockSvc,
	}

	resources, err := lister.List(context.TODO(), testListerOpts)

	a.Nil(err)
	a.Len(resources, 1)
}

func Test_Mock_AutoScalingLaunchConfiguration_Remove(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mock_autoscalingiface.NewMockAutoScalingAPI(ctrl)

	mockSvc.EXPECT().DeleteLaunchConfiguration(gomock.Eq(&autoscaling.DeleteLaunchConfigurationInput{
		LaunchConfigurationName: ptr.String("foo"),
	})).Return(&autoscaling.DeleteLaunchConfigurationOutput{}, nil)

	resource := AutoScalingLaunchConfiguration{
		svc:  mockSvc,
		Name: ptr.String("foo"),
	}

	err := resource.Remove(context.TODO())
	a.Nil(err)
}

func Test_Mock_AutoScalingLaunchConfiguration_Properties(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mock_autoscalingiface.NewMockAutoScalingAPI(ctrl)

	resource := AutoScalingLaunchConfiguration{
		svc:  mockSvc,
		Name: ptr.String("foo"),
	}

	properties := resource.Properties()

	a.Equal("foo", properties.Get("Name"))
}
