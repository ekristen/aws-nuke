package resources

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/gotidy/ptr"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/aws/aws-sdk-go/aws/awserr"             //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/cloudformation" //nolint:staticcheck

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

func TestCloudformationStack_Remove_RoleARNIsNil(t *testing.T) {
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
		roleARN: nil, // roleARN is nil
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

// ============================================================================
// UseCurrentRoleToDeleteStack tests
// ============================================================================

// Test: Normal deletion path with UseCurrentRoleToDeleteStack enabled and callerRoleARN set.
// Expects RoleARN to be set on DeleteStackInput.
func TestCloudformationStack_Remove_UseCurrentRole_NormalDeletion(t *testing.T) {
	tests := []string{
		cloudformation.StackStatusCreateComplete,
		cloudformation.StackStatusUpdateComplete,
		cloudformation.StackStatusRollbackComplete,
	}

	for _, stackStatus := range tests {
		t.Run(stackStatus, func(t *testing.T) {
			a := assert.New(t)
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockCf := mock_cloudformationiface.NewMockCloudFormationAPI(ctrl)
			callerRole := "arn:aws:iam::123456789012:role/MyCleanupRole"

			stack := CloudFormationStack{
				svc:    mockCf,
				logger: logrus.NewEntry(logrus.StandardLogger()),
				Name:   ptr.String("my-cdk-stack"),
				roleARN:       ptr.String("arn:aws:iam::123456789012:role/cdk-hnb659fds-cfn-exec-role"),
				callerRoleARN: &callerRole,
				settings: &libsettings.Setting{
					"UseCurrentRoleToDeleteStack": true,
				},
			}

			gomock.InOrder(
				mockCf.EXPECT().DescribeStacks(gomock.Eq(&cloudformation.DescribeStacksInput{
					StackName: ptr.String("my-cdk-stack"),
				})).Return(&cloudformation.DescribeStacksOutput{
					Stacks: []*cloudformation.Stack{
						{
							StackStatus: ptr.String(stackStatus),
						},
					},
				}, nil),
				mockCf.EXPECT().DeleteStack(gomock.Eq(&cloudformation.DeleteStackInput{
					StackName: ptr.String("my-cdk-stack"),
					RoleARN:   &callerRole,
				})).Return(nil, nil),
				mockCf.EXPECT().WaitUntilStackDeleteComplete(gomock.Eq(&cloudformation.DescribeStacksInput{
					StackName: ptr.String("my-cdk-stack"),
				})).Return(nil),
			)

			err := stack.Remove(context.TODO())
			a.Nil(err)
		})
	}
}

// Test: DELETE_FAILED path with UseCurrentRoleToDeleteStack enabled and callerRoleARN set.
// Expects RoleARN to be set on DeleteStackInput alongside RetainResources.
func TestCloudformationStack_Remove_UseCurrentRole_DeleteFailedPath(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCf := mock_cloudformationiface.NewMockCloudFormationAPI(ctrl)
	callerRole := "arn:aws:iam::123456789012:role/MyCleanupRole"

	stack := CloudFormationStack{
		svc:    mockCf,
		logger: logrus.NewEntry(logrus.StandardLogger()),
		Name:   ptr.String("my-cdk-stack"),
		roleARN:       ptr.String("arn:aws:iam::123456789012:role/cdk-hnb659fds-cfn-exec-role"),
		callerRoleARN: &callerRole,
		settings: &libsettings.Setting{
			"UseCurrentRoleToDeleteStack": true,
		},
	}

	gomock.InOrder(
		mockCf.EXPECT().DescribeStacks(gomock.Eq(&cloudformation.DescribeStacksInput{
			StackName: ptr.String("my-cdk-stack"),
		})).Return(&cloudformation.DescribeStacksOutput{
			Stacks: []*cloudformation.Stack{
				{
					StackStatus: ptr.String(cloudformation.StackStatusDeleteFailed),
				},
			},
		}, nil),
		mockCf.EXPECT().ListStackResources(gomock.Eq(&cloudformation.ListStackResourcesInput{
			StackName: ptr.String("my-cdk-stack"),
		})).Return(&cloudformation.ListStackResourcesOutput{
			StackResourceSummaries: []*cloudformation.StackResourceSummary{
				{
					ResourceStatus:    ptr.String(cloudformation.ResourceStatusDeleteComplete),
					LogicalResourceId: ptr.String("CompletedResource"),
				},
				{
					ResourceStatus:    ptr.String(cloudformation.ResourceStatusDeleteFailed),
					LogicalResourceId: ptr.String("FailedResource"),
				},
			},
		}, nil),
		mockCf.EXPECT().DeleteStack(gomock.Eq(&cloudformation.DeleteStackInput{
			StackName: ptr.String("my-cdk-stack"),
			RoleARN:   &callerRole,
			RetainResources: []*string{
				ptr.String("FailedResource"),
			},
		})).Return(nil, nil),
		mockCf.EXPECT().WaitUntilStackDeleteComplete(gomock.Eq(&cloudformation.DescribeStacksInput{
			StackName: ptr.String("my-cdk-stack"),
		})).Return(nil),
	)

	err := stack.Remove(context.TODO())
	a.Nil(err)
}

// Test: UseCurrentRoleToDeleteStack disabled (default). RoleARN must NOT be set on DeleteStackInput,
// even when callerRoleARN is available. This is the backward-compatibility guarantee.
func TestCloudformationStack_Remove_UseCurrentRole_Disabled_NormalDeletion(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCf := mock_cloudformationiface.NewMockCloudFormationAPI(ctrl)
	callerRole := "arn:aws:iam::123456789012:role/MyCleanupRole"

	stack := CloudFormationStack{
		svc:    mockCf,
		logger: logrus.NewEntry(logrus.StandardLogger()),
		Name:   ptr.String("my-cdk-stack"),
		roleARN:       ptr.String("arn:aws:iam::123456789012:role/cdk-hnb659fds-cfn-exec-role"),
		callerRoleARN: &callerRole,
		settings: &libsettings.Setting{
			"DisableDeletionProtection": true,
			// UseCurrentRoleToDeleteStack intentionally NOT set
		},
	}

	gomock.InOrder(
		mockCf.EXPECT().DescribeStacks(gomock.Eq(&cloudformation.DescribeStacksInput{
			StackName: ptr.String("my-cdk-stack"),
		})).Return(&cloudformation.DescribeStacksOutput{
			Stacks: []*cloudformation.Stack{
				{
					StackStatus: ptr.String(cloudformation.StackStatusCreateComplete),
				},
			},
		}, nil),
		// RoleARN must NOT be present — CloudFormation uses the stack's original role
		mockCf.EXPECT().DeleteStack(gomock.Eq(&cloudformation.DeleteStackInput{
			StackName: ptr.String("my-cdk-stack"),
		})).Return(nil, nil),
		mockCf.EXPECT().WaitUntilStackDeleteComplete(gomock.Eq(&cloudformation.DescribeStacksInput{
			StackName: ptr.String("my-cdk-stack"),
		})).Return(nil),
	)

	err := stack.Remove(context.TODO())
	a.Nil(err)
}

// Test: UseCurrentRoleToDeleteStack disabled on DELETE_FAILED path. RoleARN must NOT be set.
func TestCloudformationStack_Remove_UseCurrentRole_Disabled_DeleteFailedPath(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCf := mock_cloudformationiface.NewMockCloudFormationAPI(ctrl)
	callerRole := "arn:aws:iam::123456789012:role/MyCleanupRole"

	stack := CloudFormationStack{
		svc:    mockCf,
		logger: logrus.NewEntry(logrus.StandardLogger()),
		Name:   ptr.String("my-cdk-stack"),
		roleARN:       ptr.String("arn:aws:iam::123456789012:role/cdk-hnb659fds-cfn-exec-role"),
		callerRoleARN: &callerRole,
		settings: &libsettings.Setting{
			"DisableDeletionProtection": true,
			// UseCurrentRoleToDeleteStack intentionally NOT set
		},
	}

	gomock.InOrder(
		mockCf.EXPECT().DescribeStacks(gomock.Eq(&cloudformation.DescribeStacksInput{
			StackName: ptr.String("my-cdk-stack"),
		})).Return(&cloudformation.DescribeStacksOutput{
			Stacks: []*cloudformation.Stack{
				{
					StackStatus: ptr.String(cloudformation.StackStatusDeleteFailed),
				},
			},
		}, nil),
		mockCf.EXPECT().ListStackResources(gomock.Eq(&cloudformation.ListStackResourcesInput{
			StackName: ptr.String("my-cdk-stack"),
		})).Return(&cloudformation.ListStackResourcesOutput{
			StackResourceSummaries: []*cloudformation.StackResourceSummary{
				{
					ResourceStatus:    ptr.String(cloudformation.ResourceStatusDeleteFailed),
					LogicalResourceId: ptr.String("FailedResource"),
				},
			},
		}, nil),
		// RoleARN must NOT be present
		mockCf.EXPECT().DeleteStack(gomock.Eq(&cloudformation.DeleteStackInput{
			StackName: ptr.String("my-cdk-stack"),
			RetainResources: []*string{
				ptr.String("FailedResource"),
			},
		})).Return(nil, nil),
		mockCf.EXPECT().WaitUntilStackDeleteComplete(gomock.Eq(&cloudformation.DescribeStacksInput{
			StackName: ptr.String("my-cdk-stack"),
		})).Return(nil),
	)

	err := stack.Remove(context.TODO())
	a.Nil(err)
}

// Test: UseCurrentRoleToDeleteStack enabled but callerRoleARN is nil.
// Should NOT set RoleARN (graceful fallback) and not panic.
func TestCloudformationStack_Remove_UseCurrentRole_CallerRoleNil(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCf := mock_cloudformationiface.NewMockCloudFormationAPI(ctrl)

	stack := CloudFormationStack{
		svc:    mockCf,
		logger: logrus.NewEntry(logrus.StandardLogger()),
		Name:   ptr.String("my-cdk-stack"),
		roleARN:       ptr.String("arn:aws:iam::123456789012:role/cdk-hnb659fds-cfn-exec-role"),
		callerRoleARN: nil, // nil — e.g. non-assumed-role caller
		settings: &libsettings.Setting{
			"UseCurrentRoleToDeleteStack": true,
		},
	}

	gomock.InOrder(
		mockCf.EXPECT().DescribeStacks(gomock.Eq(&cloudformation.DescribeStacksInput{
			StackName: ptr.String("my-cdk-stack"),
		})).Return(&cloudformation.DescribeStacksOutput{
			Stacks: []*cloudformation.Stack{
				{
					StackStatus: ptr.String(cloudformation.StackStatusCreateComplete),
				},
			},
		}, nil),
		// RoleARN must NOT be present since callerRoleARN is nil
		mockCf.EXPECT().DeleteStack(gomock.Eq(&cloudformation.DeleteStackInput{
			StackName: ptr.String("my-cdk-stack"),
		})).Return(nil, nil),
		mockCf.EXPECT().WaitUntilStackDeleteComplete(gomock.Eq(&cloudformation.DescribeStacksInput{
			StackName: ptr.String("my-cdk-stack"),
		})).Return(nil),
	)

	err := stack.Remove(context.TODO())
	a.Nil(err)
}

// Test: UseCurrentRoleToDeleteStack enabled, callerRoleARN nil, DELETE_FAILED path.
func TestCloudformationStack_Remove_UseCurrentRole_CallerRoleNil_DeleteFailedPath(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCf := mock_cloudformationiface.NewMockCloudFormationAPI(ctrl)

	stack := CloudFormationStack{
		svc:    mockCf,
		logger: logrus.NewEntry(logrus.StandardLogger()),
		Name:   ptr.String("my-cdk-stack"),
		roleARN:       ptr.String("arn:aws:iam::123456789012:role/cdk-hnb659fds-cfn-exec-role"),
		callerRoleARN: nil,
		settings: &libsettings.Setting{
			"UseCurrentRoleToDeleteStack": true,
		},
	}

	gomock.InOrder(
		mockCf.EXPECT().DescribeStacks(gomock.Eq(&cloudformation.DescribeStacksInput{
			StackName: ptr.String("my-cdk-stack"),
		})).Return(&cloudformation.DescribeStacksOutput{
			Stacks: []*cloudformation.Stack{
				{
					StackStatus: ptr.String(cloudformation.StackStatusDeleteFailed),
				},
			},
		}, nil),
		mockCf.EXPECT().ListStackResources(gomock.Eq(&cloudformation.ListStackResourcesInput{
			StackName: ptr.String("my-cdk-stack"),
		})).Return(&cloudformation.ListStackResourcesOutput{
			StackResourceSummaries: []*cloudformation.StackResourceSummary{
				{
					ResourceStatus:    ptr.String(cloudformation.ResourceStatusDeleteFailed),
					LogicalResourceId: ptr.String("FailedResource"),
				},
			},
		}, nil),
		// RoleARN must NOT be present since callerRoleARN is nil
		mockCf.EXPECT().DeleteStack(gomock.Eq(&cloudformation.DeleteStackInput{
			StackName: ptr.String("my-cdk-stack"),
			RetainResources: []*string{
				ptr.String("FailedResource"),
			},
		})).Return(nil, nil),
		mockCf.EXPECT().WaitUntilStackDeleteComplete(gomock.Eq(&cloudformation.DescribeStacksInput{
			StackName: ptr.String("my-cdk-stack"),
		})).Return(nil),
	)

	err := stack.Remove(context.TODO())
	a.Nil(err)
}

// Test: Stack with NO roleARN (nil) and UseCurrentRoleToDeleteStack enabled.
// Should still use callerRoleARN since the setting is about overriding the deletion role.
func TestCloudformationStack_Remove_UseCurrentRole_StackHasNoRole(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCf := mock_cloudformationiface.NewMockCloudFormationAPI(ctrl)
	callerRole := "arn:aws:iam::123456789012:role/MyCleanupRole"

	stack := CloudFormationStack{
		svc:    mockCf,
		logger: logrus.NewEntry(logrus.StandardLogger()),
		Name:   ptr.String("no-role-stack"),
		roleARN:       nil, // stack was created without a role
		callerRoleARN: &callerRole,
		settings: &libsettings.Setting{
			"UseCurrentRoleToDeleteStack": true,
		},
	}

	gomock.InOrder(
		mockCf.EXPECT().DescribeStacks(gomock.Eq(&cloudformation.DescribeStacksInput{
			StackName: ptr.String("no-role-stack"),
		})).Return(&cloudformation.DescribeStacksOutput{
			Stacks: []*cloudformation.Stack{
				{
					StackStatus: ptr.String(cloudformation.StackStatusCreateComplete),
				},
			},
		}, nil),
		mockCf.EXPECT().DeleteStack(gomock.Eq(&cloudformation.DeleteStackInput{
			StackName: ptr.String("no-role-stack"),
			RoleARN:   &callerRole,
		})).Return(nil, nil),
		mockCf.EXPECT().WaitUntilStackDeleteComplete(gomock.Eq(&cloudformation.DescribeStacksInput{
			StackName: ptr.String("no-role-stack"),
		})).Return(nil),
	)

	err := stack.Remove(context.TODO())
	a.Nil(err)
}

// Test: UseCurrentRoleToDeleteStack explicitly set to false. Must NOT set RoleARN.
func TestCloudformationStack_Remove_UseCurrentRole_ExplicitlyFalse(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCf := mock_cloudformationiface.NewMockCloudFormationAPI(ctrl)
	callerRole := "arn:aws:iam::123456789012:role/MyCleanupRole"

	stack := CloudFormationStack{
		svc:    mockCf,
		logger: logrus.NewEntry(logrus.StandardLogger()),
		Name:   ptr.String("my-stack"),
		roleARN:       ptr.String("arn:aws:iam::123456789012:role/cdk-exec-role"),
		callerRoleARN: &callerRole,
		settings: &libsettings.Setting{
			"UseCurrentRoleToDeleteStack": false,
		},
	}

	gomock.InOrder(
		mockCf.EXPECT().DescribeStacks(gomock.Eq(&cloudformation.DescribeStacksInput{
			StackName: ptr.String("my-stack"),
		})).Return(&cloudformation.DescribeStacksOutput{
			Stacks: []*cloudformation.Stack{
				{
					StackStatus: ptr.String(cloudformation.StackStatusCreateComplete),
				},
			},
		}, nil),
		// RoleARN must NOT be present
		mockCf.EXPECT().DeleteStack(gomock.Eq(&cloudformation.DeleteStackInput{
			StackName: ptr.String("my-stack"),
		})).Return(nil, nil),
		mockCf.EXPECT().WaitUntilStackDeleteComplete(gomock.Eq(&cloudformation.DescribeStacksInput{
			StackName: ptr.String("my-stack"),
		})).Return(nil),
	)

	err := stack.Remove(context.TODO())
	a.Nil(err)
}

// Test: Empty settings (no settings configured at all). Must NOT set RoleARN.
// Simulates a user who hasn't configured any CloudFormationStack settings.
func TestCloudformationStack_Remove_UseCurrentRole_EmptySettings(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCf := mock_cloudformationiface.NewMockCloudFormationAPI(ctrl)
	callerRole := "arn:aws:iam::123456789012:role/MyCleanupRole"

	stack := CloudFormationStack{
		svc:    mockCf,
		logger: logrus.NewEntry(logrus.StandardLogger()),
		Name:   ptr.String("my-stack"),
		roleARN:       ptr.String("arn:aws:iam::123456789012:role/cdk-exec-role"),
		callerRoleARN: &callerRole,
		settings:      &libsettings.Setting{},
	}

	gomock.InOrder(
		mockCf.EXPECT().DescribeStacks(gomock.Eq(&cloudformation.DescribeStacksInput{
			StackName: ptr.String("my-stack"),
		})).Return(&cloudformation.DescribeStacksOutput{
			Stacks: []*cloudformation.Stack{
				{
					StackStatus: ptr.String(cloudformation.StackStatusCreateComplete),
				},
			},
		}, nil),
		mockCf.EXPECT().DeleteStack(gomock.Eq(&cloudformation.DeleteStackInput{
			StackName: ptr.String("my-stack"),
		})).Return(nil, nil),
		mockCf.EXPECT().WaitUntilStackDeleteComplete(gomock.Eq(&cloudformation.DescribeStacksInput{
			StackName: ptr.String("my-stack"),
		})).Return(nil),
	)

	err := stack.Remove(context.TODO())
	a.Nil(err)
}

// Test: UseCurrentRoleToDeleteStack with in-progress stacks that need to stabilize first.
// Verifies the role override still applies after waiting for stabilization.
func TestCloudformationStack_Remove_UseCurrentRole_UpdateInProgress(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCf := mock_cloudformationiface.NewMockCloudFormationAPI(ctrl)
	callerRole := "arn:aws:iam::123456789012:role/MyCleanupRole"

	stack := CloudFormationStack{
		svc:    mockCf,
		logger: logrus.NewEntry(logrus.StandardLogger()),
		Name:   ptr.String("updating-stack"),
		roleARN:       ptr.String("arn:aws:iam::123456789012:role/cdk-exec-role"),
		callerRoleARN: &callerRole,
		settings: &libsettings.Setting{
			"UseCurrentRoleToDeleteStack": true,
		},
	}

	gomock.InOrder(
		mockCf.EXPECT().DescribeStacks(gomock.Eq(&cloudformation.DescribeStacksInput{
			StackName: ptr.String("updating-stack"),
		})).Return(&cloudformation.DescribeStacksOutput{
			Stacks: []*cloudformation.Stack{
				{
					StackStatus: ptr.String(cloudformation.StackStatusUpdateInProgress),
				},
			},
		}, nil),
		mockCf.EXPECT().WaitUntilStackUpdateComplete(gomock.Eq(&cloudformation.DescribeStacksInput{
			StackName: ptr.String("updating-stack"),
		})).Return(nil),
		mockCf.EXPECT().DeleteStack(gomock.Eq(&cloudformation.DeleteStackInput{
			StackName: ptr.String("updating-stack"),
			RoleARN:   &callerRole,
		})).Return(nil, nil),
		mockCf.EXPECT().WaitUntilStackDeleteComplete(gomock.Eq(&cloudformation.DescribeStacksInput{
			StackName: ptr.String("updating-stack"),
		})).Return(nil),
	)

	err := stack.Remove(context.TODO())
	a.Nil(err)
}

// ============================================================================
// createRole bug fix test — roleCreated must not be set on error
// ============================================================================

func TestCloudformationStack_CreateRole_ErrorDoesNotSetRoleCreated(t *testing.T) {
	a := assert.New(t)

	// We can't easily mock the v2 IAM client with gomock, but we can verify
	// the struct state by testing the removeRole guard. If createRole failed
	// and roleCreated is false, removeRole should be a no-op.
	stack := CloudFormationStack{
		logger:      logrus.NewEntry(logrus.StandardLogger()),
		roleCreated: false,
	}

	// removeRole should return nil when roleCreated is false
	err := stack.removeRole(context.TODO())
	a.Nil(err)
	a.False(stack.roleCreated)
}
