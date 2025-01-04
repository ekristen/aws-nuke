package resources

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/gotidy/ptr"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cloudformation"

	libsettings "github.com/ekristen/libnuke/pkg/settings"

	"github.com/ekristen/aws-nuke/v3/mocks/mock_cloudformationiface"
)

func TestCloudformationStack_Properties(t *testing.T) {
	a := assert.New(t)

	now := time.Now()

	stack := CloudFormationStack{
		Name:            ptr.String("foobar"),
		Status:          ptr.String(cloudformation.StackStatusCreateComplete),
		CreationTime:    ptr.Time(now),
		LastUpdatedTime: ptr.Time(now),
		Tags: []*cloudformation.Tag{
			{
				Key:   ptr.String("Name"),
				Value: ptr.String("foobar"),
			},
		},
	}

	props := stack.Properties()

	a.Equal("foobar", props.Get("Name"))
	a.Equal(cloudformation.StackStatusCreateComplete, props.Get("Status"))
	a.Equal(now.Format(time.RFC3339), props.Get("CreationTime"))
	a.Equal(now.Format(time.RFC3339), props.Get("LastUpdatedTime"))
	a.Equal("foobar", props.Get("tag:Name"))

	a.Equal("foobar", stack.String())
}

func TestCloudformationStack_Remove_StackAlreadyDeleted(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCloudformation := mock_cloudformationiface.NewMockCloudFormationAPI(ctrl)

	stack := CloudFormationStack{
		svc:    mockCloudformation,
		logger: logrus.NewEntry(logrus.StandardLogger()),
		Name:   ptr.String("foobar"),
		settings: &libsettings.Setting{
			"DisableDeletionProtection": true,
		},
	}

	mockCloudformation.EXPECT().DescribeStacks(gomock.Eq(&cloudformation.DescribeStacksInput{
		StackName: ptr.String("foobar"),
	})).Return(&cloudformation.DescribeStacksOutput{
		Stacks: []*cloudformation.Stack{
			{
				StackStatus: ptr.String(cloudformation.StackStatusDeleteComplete),
			},
		},
	}, nil)

	err := stack.Remove(context.TODO())
	a.Nil(err)
}

func TestCloudformationStack_Remove_StackDoesNotExist(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCloudformation := mock_cloudformationiface.NewMockCloudFormationAPI(ctrl)

	stack := CloudFormationStack{
		svc:    mockCloudformation,
		logger: logrus.NewEntry(logrus.StandardLogger()),
		Name:   ptr.String("foobar"),
		settings: &libsettings.Setting{
			"DisableDeletionProtection": true,
		},
	}

	mockCloudformation.EXPECT().DescribeStacks(gomock.Eq(&cloudformation.DescribeStacksInput{
		StackName: ptr.String("foobar"),
	})).Return(nil, awserr.New("ValidationFailed", "Stack with id foobar does not exist", nil))

	err := stack.Remove(context.TODO())
	a.Nil(err)
}

func TestCloudformationStack_Remove_DeleteFailed(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCloudformation := mock_cloudformationiface.NewMockCloudFormationAPI(ctrl)

	stack := CloudFormationStack{
		svc:    mockCloudformation,
		logger: logrus.NewEntry(logrus.StandardLogger()),
		Name:   ptr.String("foobar"),
		settings: &libsettings.Setting{
			"DisableDeletionProtection": true,
		},
	}

	gomock.InOrder(
		mockCloudformation.EXPECT().DescribeStacks(gomock.Eq(&cloudformation.DescribeStacksInput{
			StackName: ptr.String("foobar"),
		})).Return(&cloudformation.DescribeStacksOutput{
			Stacks: []*cloudformation.Stack{
				{
					StackStatus: ptr.String(cloudformation.StackStatusDeleteFailed),
				},
			},
		}, nil),
		mockCloudformation.EXPECT().ListStackResources(gomock.Eq(&cloudformation.ListStackResourcesInput{
			StackName: ptr.String("foobar"),
		})).Return(&cloudformation.ListStackResourcesOutput{
			StackResourceSummaries: []*cloudformation.StackResourceSummary{
				{
					ResourceStatus:    ptr.String(cloudformation.ResourceStatusDeleteComplete),
					LogicalResourceId: ptr.String("fooDeleteComplete"),
				},
				{
					ResourceStatus:    ptr.String(cloudformation.ResourceStatusDeleteFailed),
					LogicalResourceId: ptr.String("fooDeleteFailed"),
				},
			},
		}, nil),
		mockCloudformation.EXPECT().DeleteStack(gomock.Eq(&cloudformation.DeleteStackInput{
			StackName: ptr.String("foobar"),
			RetainResources: []*string{
				ptr.String("fooDeleteFailed"),
			},
		})).Return(nil, nil),
		mockCloudformation.EXPECT().WaitUntilStackDeleteComplete(gomock.Eq(&cloudformation.DescribeStacksInput{
			StackName: ptr.String("foobar"),
		})).Return(nil),
	)

	err := stack.Remove(context.TODO())
	a.Nil(err)
}

// if the stack is currently in delete in progress
func TestCloudformationStack_Remove_DeleteInProgress(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCloudformation := mock_cloudformationiface.NewMockCloudFormationAPI(ctrl)

	stack := CloudFormationStack{
		svc:    mockCloudformation,
		logger: logrus.NewEntry(logrus.StandardLogger()),
		Name:   ptr.String("foobar"),
		settings: &libsettings.Setting{
			"DisableDeletionProtection": true,
		},
	}

	gomock.InOrder(
		mockCloudformation.EXPECT().DescribeStacks(gomock.Eq(&cloudformation.DescribeStacksInput{
			StackName: ptr.String("foobar"),
		})).Return(&cloudformation.DescribeStacksOutput{
			Stacks: []*cloudformation.Stack{
				{
					StackStatus: ptr.String(cloudformation.StackStatusDeleteInProgress),
				},
			},
		}, nil),

		mockCloudformation.EXPECT().WaitUntilStackDeleteComplete(gomock.Eq(&cloudformation.DescribeStacksInput{
			StackName: ptr.String("foobar"),
		})).Return(nil),
	)

	err := stack.Remove(context.TODO())
	a.Nil(err)
}

func TestCloudformationStack_Remove_Stack_InCompletedStatus(t *testing.T) {
	tests := []string{
		cloudformation.StackStatusCreateComplete,
		cloudformation.StackStatusCreateFailed,
		cloudformation.StackStatusReviewInProgress,
		cloudformation.StackStatusRollbackComplete,
		cloudformation.StackStatusRollbackFailed,
		cloudformation.StackStatusUpdateComplete,
		cloudformation.StackStatusUpdateRollbackComplete,
		cloudformation.StackStatusUpdateRollbackFailed,
	}

	for _, stackStatus := range tests {
		t.Run(stackStatus, func(t *testing.T) {
			a := assert.New(t)
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockCloudformation := mock_cloudformationiface.NewMockCloudFormationAPI(ctrl)

			stack := CloudFormationStack{
				svc:    mockCloudformation,
				logger: logrus.NewEntry(logrus.StandardLogger()),
				Name:   ptr.String("foobar"),
				settings: &libsettings.Setting{
					"DisableDeletionProtection": true,
				},
			}

			gomock.InOrder(
				mockCloudformation.EXPECT().DescribeStacks(gomock.Eq(&cloudformation.DescribeStacksInput{
					StackName: ptr.String("foobar"),
				})).Return(&cloudformation.DescribeStacksOutput{
					Stacks: []*cloudformation.Stack{
						{
							StackStatus: ptr.String(stackStatus),
						},
					},
				}, nil),

				mockCloudformation.EXPECT().DeleteStack(gomock.Eq(&cloudformation.DeleteStackInput{
					StackName: ptr.String("foobar"),
				})).Return(nil, nil),

				mockCloudformation.EXPECT().WaitUntilStackDeleteComplete(gomock.Eq(&cloudformation.DescribeStacksInput{
					StackName: ptr.String("foobar"),
				})).Return(nil),
			)

			err := stack.Remove(context.TODO())
			a.Nil(err)
		})
	}
}

func TestCloudformationStack_Remove_Stack_CreateInProgress(t *testing.T) {
	tests := []string{
		cloudformation.StackStatusCreateInProgress,
		cloudformation.StackStatusRollbackInProgress,
	}

	for _, stackStatus := range tests {
		t.Run(stackStatus, func(t *testing.T) {
			a := assert.New(t)
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockCloudformation := mock_cloudformationiface.NewMockCloudFormationAPI(ctrl)

			stack := CloudFormationStack{
				svc:    mockCloudformation,
				logger: logrus.NewEntry(logrus.StandardLogger()),
				Name:   ptr.String("foobar"),
				settings: &libsettings.Setting{
					"DisableDeletionProtection": true,
				},
			}

			gomock.InOrder(
				mockCloudformation.EXPECT().DescribeStacks(gomock.Eq(&cloudformation.DescribeStacksInput{
					StackName: ptr.String("foobar"),
				})).Return(&cloudformation.DescribeStacksOutput{
					Stacks: []*cloudformation.Stack{
						{
							StackStatus: ptr.String(stackStatus),
						},
					},
				}, nil),

				mockCloudformation.EXPECT().WaitUntilStackCreateComplete(gomock.Eq(&cloudformation.DescribeStacksInput{
					StackName: ptr.String("foobar"),
				})).Return(nil),

				mockCloudformation.EXPECT().DeleteStack(gomock.Eq(&cloudformation.DeleteStackInput{
					StackName: ptr.String("foobar"),
				})).Return(nil, nil),

				mockCloudformation.EXPECT().WaitUntilStackDeleteComplete(gomock.Eq(&cloudformation.DescribeStacksInput{
					StackName: ptr.String("foobar"),
				})).Return(nil),
			)

			err := stack.Remove(context.TODO())
			a.Nil(err)
		})
	}
}

func TestCloudformationStack_Remove_Stack_UpdateInProgress(t *testing.T) {
	tests := []string{
		cloudformation.StackStatusUpdateInProgress,
		cloudformation.StackStatusUpdateRollbackCompleteCleanupInProgress,
		cloudformation.StackStatusUpdateRollbackInProgress,
	}

	for _, stackStatus := range tests {
		t.Run(stackStatus, func(t *testing.T) {
			a := assert.New(t)
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockCloudformation := mock_cloudformationiface.NewMockCloudFormationAPI(ctrl)

			stack := CloudFormationStack{
				svc:    mockCloudformation,
				logger: logrus.NewEntry(logrus.StandardLogger()),
				Name:   ptr.String("foobar"),
				settings: &libsettings.Setting{
					"DisableDeletionProtection": true,
				},
			}

			gomock.InOrder(
				mockCloudformation.EXPECT().DescribeStacks(gomock.Eq(&cloudformation.DescribeStacksInput{
					StackName: ptr.String("foobar"),
				})).Return(&cloudformation.DescribeStacksOutput{
					Stacks: []*cloudformation.Stack{
						{
							StackStatus: ptr.String(stackStatus),
						},
					},
				}, nil),

				mockCloudformation.EXPECT().WaitUntilStackUpdateComplete(gomock.Eq(&cloudformation.DescribeStacksInput{
					StackName: ptr.String("foobar"),
				})).Return(nil),

				mockCloudformation.EXPECT().DeleteStack(gomock.Eq(&cloudformation.DeleteStackInput{
					StackName: ptr.String("foobar"),
				})).Return(nil, nil),

				mockCloudformation.EXPECT().WaitUntilStackDeleteComplete(gomock.Eq(&cloudformation.DescribeStacksInput{
					StackName: ptr.String("foobar"),
				})).Return(nil),
			)

			err := stack.Remove(context.TODO())
			a.Nil(err)
		})
	}
}
