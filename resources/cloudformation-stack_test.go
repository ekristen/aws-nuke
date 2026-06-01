package resources

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/gotidy/ptr"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/aws/aws-sdk-go/aws/awserr"             //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/cloudformation" //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/sts"            //nolint:staticcheck

	"github.com/aws/aws-sdk-go-v2/service/iam"

	libsettings "github.com/ekristen/libnuke/pkg/settings"

	"github.com/ekristen/aws-nuke/v3/mocks/mock_cloudformationiface"
	"github.com/ekristen/aws-nuke/v3/mocks/mock_stsiface"
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

// helper to create a mock STS that returns a given assumed-role ARN
func mockSTSWithAssumedRole(ctrl *gomock.Controller, arn string) *mock_stsiface.MockSTSAPI {
	mockSts := mock_stsiface.NewMockSTSAPI(ctrl)
	mockSts.EXPECT().GetCallerIdentity(gomock.Any()).Return(&sts.GetCallerIdentityOutput{
		Arn: ptr.String(arn),
	}, nil).AnyTimes()
	return mockSts
}

// Test: Normal deletion with UseCurrentRoleToDeleteStack enabled.
// STS is called lazily, role ARN is resolved and passed to DeleteStack.
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
			mockSts := mockSTSWithAssumedRole(ctrl, "arn:aws:sts::123456789012:assumed-role/MyCleanupRole/session")
			expectedRole := "arn:aws:iam::123456789012:role/MyCleanupRole"

			stack := CloudFormationStack{
				svc:    mockCf,
				stsSvc: mockSts,
				logger: logrus.NewEntry(logrus.StandardLogger()),
				Name:   ptr.String("my-cdk-stack"),
				settings: &libsettings.Setting{
					"UseCurrentRoleToDeleteStack": true,
				},
			}

			gomock.InOrder(
				mockCf.EXPECT().DescribeStacks(gomock.Eq(&cloudformation.DescribeStacksInput{
					StackName: ptr.String("my-cdk-stack"),
				})).Return(&cloudformation.DescribeStacksOutput{
					Stacks: []*cloudformation.Stack{
						{StackStatus: ptr.String(stackStatus)},
					},
				}, nil),
				mockCf.EXPECT().DeleteStack(gomock.Eq(&cloudformation.DeleteStackInput{
					StackName: ptr.String("my-cdk-stack"),
					RoleARN:   &expectedRole,
				})).Return(nil, nil),
				mockCf.EXPECT().WaitUntilStackDeleteComplete(gomock.Any()).Return(nil),
			)

			err := stack.Remove(context.TODO())
			a.Nil(err)
		})
	}
}

// Test: DELETE_FAILED path with UseCurrentRoleToDeleteStack enabled.
func TestCloudformationStack_Remove_UseCurrentRole_DeleteFailedPath(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCf := mock_cloudformationiface.NewMockCloudFormationAPI(ctrl)
	mockSts := mockSTSWithAssumedRole(ctrl, "arn:aws:sts::123456789012:assumed-role/MyCleanupRole/session")
	expectedRole := "arn:aws:iam::123456789012:role/MyCleanupRole"

	stack := CloudFormationStack{
		svc:    mockCf,
		stsSvc: mockSts,
		logger: logrus.NewEntry(logrus.StandardLogger()),
		Name:   ptr.String("my-cdk-stack"),
		settings: &libsettings.Setting{
			"UseCurrentRoleToDeleteStack": true,
		},
	}

	gomock.InOrder(
		mockCf.EXPECT().DescribeStacks(gomock.Any()).Return(&cloudformation.DescribeStacksOutput{
			Stacks: []*cloudformation.Stack{
				{StackStatus: ptr.String(cloudformation.StackStatusDeleteFailed)},
			},
		}, nil),
		mockCf.EXPECT().ListStackResources(gomock.Any()).Return(&cloudformation.ListStackResourcesOutput{
			StackResourceSummaries: []*cloudformation.StackResourceSummary{
				{ResourceStatus: ptr.String(cloudformation.ResourceStatusDeleteFailed), LogicalResourceId: ptr.String("FailedResource")},
			},
		}, nil),
		mockCf.EXPECT().DeleteStack(gomock.Eq(&cloudformation.DeleteStackInput{
			StackName:       ptr.String("my-cdk-stack"),
			RoleARN:         &expectedRole,
			RetainResources: []*string{ptr.String("FailedResource")},
		})).Return(nil, nil),
		mockCf.EXPECT().WaitUntilStackDeleteComplete(gomock.Any()).Return(nil),
	)

	err := stack.Remove(context.TODO())
	a.Nil(err)
}

// Test: UseCurrentRoleToDeleteStack disabled — STS must NOT be called, RoleARN must NOT be set.
func TestCloudformationStack_Remove_UseCurrentRole_Disabled(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCf := mock_cloudformationiface.NewMockCloudFormationAPI(ctrl)
	// STS mock with zero expected calls — verifies STS is never called when setting is disabled
	mockSts := mock_stsiface.NewMockSTSAPI(ctrl)

	stack := CloudFormationStack{
		svc:    mockCf,
		stsSvc: mockSts,
		logger: logrus.NewEntry(logrus.StandardLogger()),
		Name:   ptr.String("my-stack"),
		settings: &libsettings.Setting{
			"DisableDeletionProtection": true,
		},
	}

	gomock.InOrder(
		mockCf.EXPECT().DescribeStacks(gomock.Any()).Return(&cloudformation.DescribeStacksOutput{
			Stacks: []*cloudformation.Stack{
				{StackStatus: ptr.String(cloudformation.StackStatusCreateComplete)},
			},
		}, nil),
		mockCf.EXPECT().DeleteStack(gomock.Eq(&cloudformation.DeleteStackInput{
			StackName: ptr.String("my-stack"),
		})).Return(nil, nil),
		mockCf.EXPECT().WaitUntilStackDeleteComplete(gomock.Any()).Return(nil),
	)

	err := stack.Remove(context.TODO())
	a.Nil(err)
}

// Test: UseCurrentRoleToDeleteStack disabled on DELETE_FAILED path.
func TestCloudformationStack_Remove_UseCurrentRole_Disabled_DeleteFailedPath(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCf := mock_cloudformationiface.NewMockCloudFormationAPI(ctrl)
	mockSts := mock_stsiface.NewMockSTSAPI(ctrl) // no expected calls

	stack := CloudFormationStack{
		svc:    mockCf,
		stsSvc: mockSts,
		logger: logrus.NewEntry(logrus.StandardLogger()),
		Name:   ptr.String("my-stack"),
		settings: &libsettings.Setting{
			"DisableDeletionProtection": true,
		},
	}

	gomock.InOrder(
		mockCf.EXPECT().DescribeStacks(gomock.Any()).Return(&cloudformation.DescribeStacksOutput{
			Stacks: []*cloudformation.Stack{
				{StackStatus: ptr.String(cloudformation.StackStatusDeleteFailed)},
			},
		}, nil),
		mockCf.EXPECT().ListStackResources(gomock.Any()).Return(&cloudformation.ListStackResourcesOutput{
			StackResourceSummaries: []*cloudformation.StackResourceSummary{
				{ResourceStatus: ptr.String(cloudformation.ResourceStatusDeleteFailed), LogicalResourceId: ptr.String("Res")},
			},
		}, nil),
		mockCf.EXPECT().DeleteStack(gomock.Eq(&cloudformation.DeleteStackInput{
			StackName:       ptr.String("my-stack"),
			RetainResources: []*string{ptr.String("Res")},
		})).Return(nil, nil),
		mockCf.EXPECT().WaitUntilStackDeleteComplete(gomock.Any()).Return(nil),
	)

	err := stack.Remove(context.TODO())
	a.Nil(err)
}

// Test: STS GetCallerIdentity fails — should log warning and fall back to default behavior.
func TestCloudformationStack_Remove_UseCurrentRole_STSError(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCf := mock_cloudformationiface.NewMockCloudFormationAPI(ctrl)
	mockSts := mock_stsiface.NewMockSTSAPI(ctrl)
	mockSts.EXPECT().GetCallerIdentity(gomock.Any()).Return(nil, fmt.Errorf("access denied")).AnyTimes()

	stack := CloudFormationStack{
		svc:    mockCf,
		stsSvc: mockSts,
		logger: logrus.NewEntry(logrus.StandardLogger()),
		Name:   ptr.String("my-stack"),
		settings: &libsettings.Setting{
			"UseCurrentRoleToDeleteStack": true,
		},
	}

	gomock.InOrder(
		mockCf.EXPECT().DescribeStacks(gomock.Any()).Return(&cloudformation.DescribeStacksOutput{
			Stacks: []*cloudformation.Stack{
				{StackStatus: ptr.String(cloudformation.StackStatusCreateComplete)},
			},
		}, nil),
		// RoleARN must NOT be set since STS failed
		mockCf.EXPECT().DeleteStack(gomock.Eq(&cloudformation.DeleteStackInput{
			StackName: ptr.String("my-stack"),
		})).Return(nil, nil),
		mockCf.EXPECT().WaitUntilStackDeleteComplete(gomock.Any()).Return(nil),
	)

	err := stack.Remove(context.TODO())
	a.Nil(err)
}

// Test: Caller is not using an assumed role (e.g. IAM user) — should log warning and fall back.
func TestCloudformationStack_Remove_UseCurrentRole_NotAssumedRole(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCf := mock_cloudformationiface.NewMockCloudFormationAPI(ctrl)
	mockSts := mock_stsiface.NewMockSTSAPI(ctrl)
	mockSts.EXPECT().GetCallerIdentity(gomock.Any()).Return(&sts.GetCallerIdentityOutput{
		Arn: ptr.String("arn:aws:iam::123456789012:user/MyUser"),
	}, nil).AnyTimes()

	stack := CloudFormationStack{
		svc:    mockCf,
		stsSvc: mockSts,
		logger: logrus.NewEntry(logrus.StandardLogger()),
		Name:   ptr.String("my-stack"),
		settings: &libsettings.Setting{
			"UseCurrentRoleToDeleteStack": true,
		},
	}

	gomock.InOrder(
		mockCf.EXPECT().DescribeStacks(gomock.Any()).Return(&cloudformation.DescribeStacksOutput{
			Stacks: []*cloudformation.Stack{
				{StackStatus: ptr.String(cloudformation.StackStatusCreateComplete)},
			},
		}, nil),
		mockCf.EXPECT().DeleteStack(gomock.Eq(&cloudformation.DeleteStackInput{
			StackName: ptr.String("my-stack"),
		})).Return(nil, nil),
		mockCf.EXPECT().WaitUntilStackDeleteComplete(gomock.Any()).Return(nil),
	)

	err := stack.Remove(context.TODO())
	a.Nil(err)
}

// Test: GovCloud partition — ARN reconstruction must use aws-us-gov, not aws.
func TestCloudformationStack_Remove_UseCurrentRole_GovCloudPartition(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCf := mock_cloudformationiface.NewMockCloudFormationAPI(ctrl)
	mockSts := mockSTSWithAssumedRole(ctrl, "arn:aws-us-gov:sts::123456789012:assumed-role/GovRole/session")
	expectedRole := "arn:aws-us-gov:iam::123456789012:role/GovRole"

	stack := CloudFormationStack{
		svc:    mockCf,
		stsSvc: mockSts,
		logger: logrus.NewEntry(logrus.StandardLogger()),
		Name:   ptr.String("gov-stack"),
		settings: &libsettings.Setting{
			"UseCurrentRoleToDeleteStack": true,
		},
	}

	gomock.InOrder(
		mockCf.EXPECT().DescribeStacks(gomock.Any()).Return(&cloudformation.DescribeStacksOutput{
			Stacks: []*cloudformation.Stack{
				{StackStatus: ptr.String(cloudformation.StackStatusCreateComplete)},
			},
		}, nil),
		mockCf.EXPECT().DeleteStack(gomock.Eq(&cloudformation.DeleteStackInput{
			StackName: ptr.String("gov-stack"),
			RoleARN:   &expectedRole,
		})).Return(nil, nil),
		mockCf.EXPECT().WaitUntilStackDeleteComplete(gomock.Any()).Return(nil),
	)

	err := stack.Remove(context.TODO())
	a.Nil(err)
}

// Test: China partition — ARN reconstruction must use aws-cn.
func TestCloudformationStack_Remove_UseCurrentRole_ChinaPartition(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCf := mock_cloudformationiface.NewMockCloudFormationAPI(ctrl)
	mockSts := mockSTSWithAssumedRole(ctrl, "arn:aws-cn:sts::123456789012:assumed-role/ChinaRole/session")
	expectedRole := "arn:aws-cn:iam::123456789012:role/ChinaRole"

	stack := CloudFormationStack{
		svc:    mockCf,
		stsSvc: mockSts,
		logger: logrus.NewEntry(logrus.StandardLogger()),
		Name:   ptr.String("cn-stack"),
		settings: &libsettings.Setting{
			"UseCurrentRoleToDeleteStack": true,
		},
	}

	gomock.InOrder(
		mockCf.EXPECT().DescribeStacks(gomock.Any()).Return(&cloudformation.DescribeStacksOutput{
			Stacks: []*cloudformation.Stack{
				{StackStatus: ptr.String(cloudformation.StackStatusCreateComplete)},
			},
		}, nil),
		mockCf.EXPECT().DeleteStack(gomock.Eq(&cloudformation.DeleteStackInput{
			StackName: ptr.String("cn-stack"),
			RoleARN:   &expectedRole,
		})).Return(nil, nil),
		mockCf.EXPECT().WaitUntilStackDeleteComplete(gomock.Any()).Return(nil),
	)

	err := stack.Remove(context.TODO())
	a.Nil(err)
}

// Test: STS is only called once even across multiple deletion attempts (lazy + cached).
func TestCloudformationStack_Remove_UseCurrentRole_STSCalledOnce(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCf := mock_cloudformationiface.NewMockCloudFormationAPI(ctrl)
	mockSts := mock_stsiface.NewMockSTSAPI(ctrl)
	// Expect exactly 1 STS call even though doRemove is called twice (retry)
	mockSts.EXPECT().GetCallerIdentity(gomock.Any()).Return(&sts.GetCallerIdentityOutput{
		Arn: ptr.String("arn:aws:sts::123456789012:assumed-role/MyRole/session"),
	}, nil).Times(1)

	expectedRole := "arn:aws:iam::123456789012:role/MyRole"

	stack := CloudFormationStack{
		svc:               mockCf,
		stsSvc:            mockSts,
		logger:            logrus.NewEntry(logrus.StandardLogger()),
		Name:              ptr.String("retry-stack"),
		maxDeleteAttempts: 3,
		settings: &libsettings.Setting{
			"UseCurrentRoleToDeleteStack": true,
		},
	}

	// First attempt: DeleteStack fails
	mockCf.EXPECT().DescribeStacks(gomock.Any()).Return(&cloudformation.DescribeStacksOutput{
		Stacks: []*cloudformation.Stack{
			{StackStatus: ptr.String(cloudformation.StackStatusCreateComplete)},
		},
	}, nil).Times(2)

	firstCall := mockCf.EXPECT().DeleteStack(gomock.Eq(&cloudformation.DeleteStackInput{
		StackName: ptr.String("retry-stack"),
		RoleARN:   &expectedRole,
	})).Return(nil, awserr.New("InternalError", "transient failure", nil))

	mockCf.EXPECT().DeleteStack(gomock.Eq(&cloudformation.DeleteStackInput{
		StackName: ptr.String("retry-stack"),
		RoleARN:   &expectedRole,
	})).After(firstCall).Return(nil, nil)

	mockCf.EXPECT().WaitUntilStackDeleteComplete(gomock.Any()).Return(nil)

	err := stack.Remove(context.TODO())
	a.Nil(err)
}

// Test: Empty settings — STS must NOT be called.
func TestCloudformationStack_Remove_UseCurrentRole_EmptySettings(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCf := mock_cloudformationiface.NewMockCloudFormationAPI(ctrl)
	mockSts := mock_stsiface.NewMockSTSAPI(ctrl) // no expected calls

	stack := CloudFormationStack{
		svc:      mockCf,
		stsSvc:   mockSts,
		logger:   logrus.NewEntry(logrus.StandardLogger()),
		Name:     ptr.String("my-stack"),
		settings: &libsettings.Setting{},
	}

	gomock.InOrder(
		mockCf.EXPECT().DescribeStacks(gomock.Any()).Return(&cloudformation.DescribeStacksOutput{
			Stacks: []*cloudformation.Stack{
				{StackStatus: ptr.String(cloudformation.StackStatusCreateComplete)},
			},
		}, nil),
		mockCf.EXPECT().DeleteStack(gomock.Eq(&cloudformation.DeleteStackInput{
			StackName: ptr.String("my-stack"),
		})).Return(nil, nil),
		mockCf.EXPECT().WaitUntilStackDeleteComplete(gomock.Any()).Return(nil),
	)

	err := stack.Remove(context.TODO())
	a.Nil(err)
}

// Test: UseCurrentRoleToDeleteStack with in-progress stack that needs stabilization.
func TestCloudformationStack_Remove_UseCurrentRole_UpdateInProgress(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCf := mock_cloudformationiface.NewMockCloudFormationAPI(ctrl)
	mockSts := mockSTSWithAssumedRole(ctrl, "arn:aws:sts::123456789012:assumed-role/MyRole/session")
	expectedRole := "arn:aws:iam::123456789012:role/MyRole"

	stack := CloudFormationStack{
		svc:    mockCf,
		stsSvc: mockSts,
		logger: logrus.NewEntry(logrus.StandardLogger()),
		Name:   ptr.String("updating-stack"),
		settings: &libsettings.Setting{
			"UseCurrentRoleToDeleteStack": true,
		},
	}

	gomock.InOrder(
		mockCf.EXPECT().DescribeStacks(gomock.Any()).Return(&cloudformation.DescribeStacksOutput{
			Stacks: []*cloudformation.Stack{
				{StackStatus: ptr.String(cloudformation.StackStatusUpdateInProgress)},
			},
		}, nil),
		mockCf.EXPECT().WaitUntilStackUpdateComplete(gomock.Any()).Return(nil),
		mockCf.EXPECT().DeleteStack(gomock.Eq(&cloudformation.DeleteStackInput{
			StackName: ptr.String("updating-stack"),
			RoleARN:   &expectedRole,
		})).Return(nil, nil),
		mockCf.EXPECT().WaitUntilStackDeleteComplete(gomock.Any()).Return(nil),
	)

	err := stack.Remove(context.TODO())
	a.Nil(err)
}

// ============================================================================
// createRole bug fix test
// ============================================================================

// fakeIAMRoleAPI is a minimal test double for the iamRoleAPI interface.
type fakeIAMRoleAPI struct {
	createRoleErr error
	deleteRoleErr error
}

func (f *fakeIAMRoleAPI) CreateRole(_ context.Context, _ *iam.CreateRoleInput, _ ...func(*iam.Options)) (*iam.CreateRoleOutput, error) {
	return nil, f.createRoleErr
}

func (f *fakeIAMRoleAPI) DeleteRole(_ context.Context, _ *iam.DeleteRoleInput, _ ...func(*iam.Options)) (*iam.DeleteRoleOutput, error) {
	return nil, f.deleteRoleErr
}

// Test: When CreateRole fails, roleCreated must remain false so removeRole is a no-op.
func TestCloudformationStack_CreateRole_ErrorDoesNotSetRoleCreated(t *testing.T) {
	a := assert.New(t)

	stack := CloudFormationStack{
		logger:  logrus.NewEntry(logrus.StandardLogger()),
		roleARN: ptr.String("arn:aws:iam::123456789012:role/SomeRole"),
		iamSvc:  &fakeIAMRoleAPI{createRoleErr: fmt.Errorf("AccessDenied: not authorized")},
	}

	err := stack.createRole(context.TODO())
	a.NotNil(err)
	a.Contains(err.Error(), "AccessDenied")
	a.False(stack.roleCreated, "roleCreated must stay false when CreateRole fails")
	a.Empty(stack.roleName, "roleName must stay empty when CreateRole fails")

	// Confirm removeRole is a safe no-op after a failed createRole
	err = stack.removeRole(context.TODO())
	a.Nil(err)
}

// Test: When CreateRole succeeds, roleCreated is set and roleName is populated.
func TestCloudformationStack_CreateRole_SuccessSetsRoleCreated(t *testing.T) {
	a := assert.New(t)

	stack := CloudFormationStack{
		logger:  logrus.NewEntry(logrus.StandardLogger()),
		roleARN: ptr.String("arn:aws:iam::123456789012:role/SomeRole"),
		iamSvc:  &fakeIAMRoleAPI{},
	}

	err := stack.createRole(context.TODO())
	a.Nil(err)
	a.True(stack.roleCreated, "roleCreated must be true after successful CreateRole")
	a.Equal("SomeRole", stack.roleName)
}
