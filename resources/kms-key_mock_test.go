package resources

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"

	"github.com/ekristen/aws-nuke/v3/mocks/mock_kmsiface"
	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

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

	lister := KMSKeyLister{
		mockSvc: mockKMS,
	}

	resources, err := lister.List(context.TODO(), &nuke.ListerOpts{
		Region: &nuke.Region{
			Name: "us-east-2",
		},
		Session: session.Must(session.NewSession()),
	})
	a.NoError(err)
	a.Len(resources, 1)
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

	lister := KMSKeyLister{
		mockSvc: mockKMS,
	}

	resources, err := lister.List(context.TODO(), &nuke.ListerOpts{
		Region: &nuke.Region{
			Name: "us-east-2",
		},
		Session: session.Must(session.NewSession()),
	})
	a.NoError(err)
	a.Len(resources, 1)
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
		Tags: []*kms.Tag{
			{TagKey: aws.String("Environment"), TagValue: aws.String("Test")},
		},
	}

	assert.Equal(t, "test-key-id", kmsKey.String())
	assert.Equal(t, kms.KeyStateEnabled, kmsKey.Properties().Get("State"))
	assert.Equal(t, kms.KeyManagerTypeCustomer, kmsKey.Properties().Get("Manager"))
	assert.Equal(t, "Test", kmsKey.Properties().Get("tag:Environment"))
}

func Test_Mock_KMSKey_Remove(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockKMS := mock_kmsiface.NewMockKMSAPI(ctrl)

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
