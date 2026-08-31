package resources

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"

	"github.com/aws/aws-sdk-go/aws"         //nolint:staticcheck
	"github.com/aws/aws-sdk-go/aws/awserr"  //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/kms" //nolint:staticcheck

	liberrors "github.com/ekristen/libnuke/pkg/errors"
	libsettings "github.com/ekristen/libnuke/pkg/settings"

	"github.com/ekristen/aws-nuke/v3/mocks/mock_kmsiface"
)

// mockListGrants sets up a ListGrantsPages expectation returning the supplied grantee principals.
func mockListGrants(mockKMS *mock_kmsiface.MockKMSAPI, keyID string, principals ...string) {
	grants := make([]*kms.GrantListEntry, 0, len(principals))
	for _, principal := range principals {
		grants = append(grants, &kms.GrantListEntry{GranteePrincipal: aws.String(principal)})
	}

	mockKMS.EXPECT().ListGrantsPages(&kms.ListGrantsInput{KeyId: aws.String(keyID)}, gomock.Any()).DoAndReturn(
		func(input *kms.ListGrantsInput, fn func(*kms.ListGrantsResponse, bool) bool) error {
			fn(&kms.ListGrantsResponse{Grants: grants}, true)
			return nil
		},
	)
}

func Test_Mock_KMSKey_List(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockKMS := mock_kmsiface.NewMockKMSAPI(ctrl)

	mockKMS.EXPECT().ListKeysPages(gomock.Any(), gomock.Any()).DoAndReturn(
		func(input *kms.ListKeysInput, fn func(*kms.ListKeysOutput, bool) bool) error {
			fn(&kms.ListKeysOutput{
				Keys: []*kms.KeyListEntry{
					{KeyId: aws.String("test-key-id")},
				},
			}, true)
			return nil
		},
	)

	mockKMS.EXPECT().DescribeKey(gomock.Any()).DoAndReturn(
		func(input *kms.DescribeKeyInput) (*kms.DescribeKeyOutput, error) {
			return &kms.DescribeKeyOutput{
				KeyMetadata: &kms.KeyMetadata{
					KeyId:      aws.String("test-key-id"),
					Arn:        aws.String("arn:aws:kms:us-east-2:123456789012:key/test-key-id"),
					KeyManager: aws.String(kms.KeyManagerTypeCustomer),
					KeyState:   aws.String(kms.KeyStateEnabled),
				},
			}, nil
		},
	)

	mockKMS.EXPECT().ListResourceTags(gomock.Any()).DoAndReturn(
		func(input *kms.ListResourceTagsInput) (*kms.ListResourceTagsOutput, error) {
			return &kms.ListResourceTagsOutput{
				Tags: []*kms.Tag{
					{TagKey: aws.String("Environment"), TagValue: aws.String("Test")},
				},
			}, nil
		},
	)

	mockKMS.EXPECT().ListAliases(&kms.ListAliasesInput{
		KeyId: aws.String("test-key-id"),
	}).Return(&kms.ListAliasesOutput{
		Aliases: []*kms.AliasListEntry{
			{AliasName: aws.String("alias/test-key-id")},
		},
	}, nil)

	mockListGrants(mockKMS, "test-key-id", "dynamodb.us-east-2.amazonaws.com")

	lister := KMSKeyLister{
		mockSvc: mockKMS,
	}

	resources, err := lister.List(context.TODO(), testListerOpts)
	a.NoError(err)
	a.Len(resources, 1)

	key := resources[0].(*KMSKey)
	a.True(ptr.ToBool(key.InUse))
	a.Equal("dynamodb.us-east-2.amazonaws.com", ptr.ToString(key.InUseBy))
}

func Test_Mock_KMSKey_List_WithAccessDenied(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockKMS := mock_kmsiface.NewMockKMSAPI(ctrl)

	// Mock the ListKeysPages method to return two keys
	mockKMS.EXPECT().ListKeysPages(gomock.Any(), gomock.Any()).DoAndReturn(
		func(input *kms.ListKeysInput, fn func(*kms.ListKeysOutput, bool) bool) error {
			fn(&kms.ListKeysOutput{
				Keys: []*kms.KeyListEntry{
					{KeyId: aws.String("test-key-id-1")},
					{KeyId: aws.String("test-key-id-2")},
				},
			}, true)
			return nil
		},
	)

	// Mock DescribeKey for the first key to return a valid response
	mockKMS.EXPECT().DescribeKey(&kms.DescribeKeyInput{
		KeyId: aws.String("test-key-id-1"),
	}).DoAndReturn(
		func(input *kms.DescribeKeyInput) (*kms.DescribeKeyOutput, error) {
			return &kms.DescribeKeyOutput{
				KeyMetadata: &kms.KeyMetadata{
					KeyId:      aws.String("test-key-id-1"),
					Arn:        aws.String("arn:aws:kms:us-east-2:123456789012:key/test-key-id-1"),
					KeyManager: aws.String(kms.KeyManagerTypeCustomer),
					KeyState:   aws.String(kms.KeyStateEnabled),
				},
			}, nil
		},
	)

	// Mock DescribeKey for the second key to return AccessDeniedException
	mockKMS.EXPECT().DescribeKey(&kms.DescribeKeyInput{
		KeyId: aws.String("test-key-id-2"),
	}).DoAndReturn(
		func(input *kms.DescribeKeyInput) (*kms.DescribeKeyOutput, error) {
			return nil, awserr.New("AccessDeniedException", "no resource-based policy allows the kms:DescribeKey action", nil)
		},
	)

	// Mock ListResourceTags for the first key
	mockKMS.EXPECT().ListResourceTags(&kms.ListResourceTagsInput{
		KeyId: aws.String("test-key-id-1"),
	}).DoAndReturn(
		func(input *kms.ListResourceTagsInput) (*kms.ListResourceTagsOutput, error) {
			return &kms.ListResourceTagsOutput{
				Tags: []*kms.Tag{
					{TagKey: aws.String("Environment"), TagValue: aws.String("Test")},
				},
			}, nil
		},
	)

	mockKMS.EXPECT().ListAliases(&kms.ListAliasesInput{
		KeyId: aws.String("test-key-id-1"),
	}).Return(&kms.ListAliasesOutput{
		Aliases: []*kms.AliasListEntry{
			{AliasName: aws.String("alias/test-key-id-1")},
		},
	}, nil)

	mockListGrants(mockKMS, "test-key-id-1")

	lister := KMSKeyLister{
		mockSvc: mockKMS,
	}

	resources, err := lister.List(context.TODO(), testListerOpts)
	a.NoError(err)
	a.Len(resources, 1)

	key := resources[0].(*KMSKey)
	a.False(ptr.ToBool(key.InUse))
}

func Test_Mock_KMSKey_Filter(t *testing.T) {
	cases := []struct {
		name    string
		state   string
		manager string
		error   string
	}{
		{
			name:    "aws-managed-key",
			state:   kms.KeyStateEnabled,
			manager: kms.KeyManagerTypeAws,
			error:   "cannot delete AWS managed key",
		},
		{
			name:    "pending-deletion-key",
			state:   kms.KeyStatePendingDeletion,
			manager: kms.KeyManagerTypeCustomer,
			error:   "is already in PendingDeletion state",
		},
		{
			name:    "pending-replica-deletion-key",
			state:   kms.KeyStatePendingReplicaDeletion,
			manager: kms.KeyManagerTypeCustomer,
			error:   "is already in PendingReplicaDeletion state",
		},
		{
			name:    "enabled-key",
			state:   kms.KeyStateEnabled,
			manager: kms.KeyManagerTypeCustomer,
			error:   "",
		},
	}

	for _, tc := range cases {
		kmsKey := KMSKey{
			ID:      ptr.String("test-key-id"),
			State:   ptr.String(tc.state),
			Manager: ptr.String(tc.manager),
		}

		err := kmsKey.Filter()
		if tc.error == "" {
			assert.NoError(t, err)
		} else {
			assert.ErrorContains(t, err, tc.error)
		}
	}
}

func Test_Mock_KMSKey_Properties(t *testing.T) {
	kmsKey := KMSKey{
		ID:      ptr.String("test-key-id"),
		State:   ptr.String(kms.KeyStateEnabled),
		Manager: ptr.String(kms.KeyManagerTypeCustomer),
		Alias:   ptr.String("alias/test-key-id"),
		InUse:   ptr.Bool(true),
		InUseBy: ptr.String("rds.us-east-2.amazonaws.com"),
		Tags: []*kms.Tag{
			{TagKey: aws.String("Environment"), TagValue: aws.String("Test")},
		},
	}

	assert.Equal(t, "test-key-id", kmsKey.String())
	assert.Equal(t, kms.KeyStateEnabled, kmsKey.Properties().Get("State"))
	assert.Equal(t, kms.KeyManagerTypeCustomer, kmsKey.Properties().Get("Manager"))
	assert.Equal(t, "Test", kmsKey.Properties().Get("tag:Environment"))
	assert.Equal(t, "alias/test-key-id", kmsKey.Properties().Get("Alias"))
	assert.Equal(t, "true", kmsKey.Properties().Get("InUse"))
	assert.Equal(t, "rds.us-east-2.amazonaws.com", kmsKey.Properties().Get("InUseBy"))
}

func Test_Mock_KMSKey_Remove(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockKMS := mock_kmsiface.NewMockKMSAPI(ctrl)

	mockListGrants(mockKMS, "test-key-id")

	mockKMS.EXPECT().ScheduleKeyDeletion(&kms.ScheduleKeyDeletionInput{
		KeyId:               aws.String("test-key-id"),
		PendingWindowInDays: aws.Int64(7),
	}).Return(&kms.ScheduleKeyDeletionOutput{}, nil)

	kmsKey := KMSKey{
		svc:     mockKMS,
		ID:      ptr.String("test-key-id"),
		State:   ptr.String(kms.KeyStateEnabled),
		Manager: ptr.String(kms.KeyManagerTypeCustomer),
	}

	err := kmsKey.Remove(context.TODO())
	a.NoError(err)
}

// Test_Mock_KMSKey_Remove_InUse asserts that a key an AWS service still holds a grant on is held rather than
// scheduled for deletion, so that it is retried later in the same run.
func Test_Mock_KMSKey_Remove_InUse(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockKMS := mock_kmsiface.NewMockKMSAPI(ctrl)

	mockListGrants(mockKMS, "test-key-id",
		"rds.us-east-2.amazonaws.com", "dynamodb.us-east-2.amazonaws.com")

	kmsKey := KMSKey{
		svc: mockKMS,
		ID:  ptr.String("test-key-id"),
	}

	err := kmsKey.Remove(context.TODO())

	var holdErr liberrors.ErrHoldResource
	a.ErrorAs(err, &holdErr)
	a.Equal("in use by dynamodb.us-east-2.amazonaws.com, rds.us-east-2.amazonaws.com", err.Error())
}

// Test_Mock_KMSKey_Remove_InUse_DeleteInUseKeys asserts the DeleteInUseKeys setting bypasses the in-use check
// entirely, including the ListGrants call.
func Test_Mock_KMSKey_Remove_InUse_DeleteInUseKeys(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockKMS := mock_kmsiface.NewMockKMSAPI(ctrl)

	mockKMS.EXPECT().ScheduleKeyDeletion(&kms.ScheduleKeyDeletionInput{
		KeyId:               aws.String("test-key-id"),
		PendingWindowInDays: aws.Int64(7),
	}).Return(&kms.ScheduleKeyDeletionOutput{}, nil)

	setting := libsettings.Setting{"DeleteInUseKeys": true}

	kmsKey := KMSKey{
		svc: mockKMS,
		ID:  ptr.String("test-key-id"),
	}
	kmsKey.Settings(&setting)

	a.NoError(kmsKey.Remove(context.TODO()))
}

// Test_Mock_KMSKey_Remove_IgnoresIAMGrants asserts grants held by IAM principals do not hold the key, since they are
// not tied to the lifecycle of any resource and would block deletion indefinitely.
func Test_Mock_KMSKey_Remove_IgnoresIAMGrants(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockKMS := mock_kmsiface.NewMockKMSAPI(ctrl)

	mockListGrants(mockKMS, "test-key-id", "arn:aws:iam::123456789012:role/some-role")

	mockKMS.EXPECT().ScheduleKeyDeletion(&kms.ScheduleKeyDeletionInput{
		KeyId:               aws.String("test-key-id"),
		PendingWindowInDays: aws.Int64(7),
	}).Return(&kms.ScheduleKeyDeletionOutput{}, nil)

	kmsKey := KMSKey{
		svc: mockKMS,
		ID:  ptr.String("test-key-id"),
	}

	a.NoError(kmsKey.Remove(context.TODO()))
}

// Test_Mock_KMSKey_Remove_ListGrantsError asserts that an unreadable grant list does not block deletion, preserving
// the behavior that existed before the in-use check was added.
func Test_Mock_KMSKey_Remove_ListGrantsError(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockKMS := mock_kmsiface.NewMockKMSAPI(ctrl)

	mockKMS.EXPECT().ListGrantsPages(gomock.Any(), gomock.Any()).
		Return(awserr.New("AccessDeniedException", "no permission to list grants", nil))

	mockKMS.EXPECT().ScheduleKeyDeletion(&kms.ScheduleKeyDeletionInput{
		KeyId:               aws.String("test-key-id"),
		PendingWindowInDays: aws.Int64(7),
	}).Return(&kms.ScheduleKeyDeletionOutput{}, nil)

	kmsKey := KMSKey{
		svc: mockKMS,
		ID:  ptr.String("test-key-id"),
	}

	a.NoError(kmsKey.Remove(context.TODO()))
}
