package resources

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"

	"github.com/ekristen/aws-nuke/v3/mocks/mock_autoscalingiface"
	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

func Test_Mock_AutoScalingLifeCycleHook_List(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mock_autoscalingiface.NewMockAutoScalingAPI(ctrl)

	mockSvc.EXPECT().DescribeAutoScalingGroups(gomock.Eq(&autoscaling.DescribeAutoScalingGroupsInput{})).
		Return(&autoscaling.DescribeAutoScalingGroupsOutput{
			AutoScalingGroups: []*autoscaling.Group{
				{
					AutoScalingGroupName: aws.String("foobar-group"),
				},
			},
		}, nil)

	mockSvc.EXPECT().DescribeLifecycleHooks(gomock.Eq(&autoscaling.DescribeLifecycleHooksInput{
		AutoScalingGroupName: aws.String("foobar-group"),
	})).
		Return(&autoscaling.DescribeLifecycleHooksOutput{
			LifecycleHooks: []*autoscaling.LifecycleHook{
				{
					LifecycleHookName:    aws.String("foobar-hook"),
					AutoScalingGroupName: aws.String("foobar-group"),
				},
			},
		}, nil)

	lister := AutoScalingLifecycleHookLister{
		mockSvc: mockSvc,
	}

	resources, err := lister.List(context.TODO(), &nuke.ListerOpts{
		Region: &nuke.Region{
			Name: "us-east-2",
		},
		Session: session.Must(session.NewSession(&aws.Config{})),
	})

	a.NoError(err)
	a.Len(resources, 1)
}

func Test_Mock_AutoScalingLifeCycleHook_Remove(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mock_autoscalingiface.NewMockAutoScalingAPI(ctrl)

	mockSvc.EXPECT().DeleteLifecycleHook(gomock.Eq(&autoscaling.DeleteLifecycleHookInput{
		AutoScalingGroupName: aws.String("foobar-group"),
		LifecycleHookName:    aws.String("foobar-hook"),
	})).Return(&autoscaling.DeleteLifecycleHookOutput{}, nil)

	resource := AutoScalingLifecycleHook{
		svc:       mockSvc,
		Name:      aws.String("foobar-hook"),
		GroupName: aws.String("foobar-group"),
	}

	err := resource.Remove(context.TODO())
	a.NoError(err)
}

func Test_Mock_AutoScalingLifeCycleHook_Properties(t *testing.T) {
	a := assert.New(t)

	resource := AutoScalingLifecycleHook{
		Name:      aws.String("foobar-hook"),
		GroupName: aws.String("foobar-group"),
	}

	props := resource.Properties()

	a.Equal("foobar-hook", props.Get("Name"))
	a.Equal("foobar-group", props.Get("GroupName"))
}
