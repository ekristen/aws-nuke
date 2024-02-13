package resources

import (
	"context"

	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const KMSKeyResource = "KMSKey"

func init() {
	registry.Register(&registry.Registration{
		Name:   KMSKeyResource,
		Scope:  nuke.Account,
		Lister: &KMSKeyLister{},
		DependsOn: []string{
			KMSAliasResource,
		},
	})
}

type KMSKeyLister struct{}

func (l *KMSKeyLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := kms.New(opts.Session)
	resources := make([]resource.Resource, 0)

	var innerErr error
	if err := svc.ListKeysPages(nil, func(resp *kms.ListKeysOutput, lastPage bool) bool {
		for _, key := range resp.Keys {
			resp, err := svc.DescribeKey(&kms.DescribeKeyInput{
				KeyId: key.KeyId,
			})
			if err != nil {
				innerErr = err
				return false
			}

			if *resp.KeyMetadata.KeyManager == kms.KeyManagerTypeAws {
				continue
			}

			if *resp.KeyMetadata.KeyState == kms.KeyStatePendingDeletion {
				continue
			}

			kmsKey := &KMSKey{
				svc:     svc,
				id:      *resp.KeyMetadata.KeyId,
				state:   *resp.KeyMetadata.KeyState,
				manager: resp.KeyMetadata.KeyManager,
			}

			tags, err := svc.ListResourceTags(&kms.ListResourceTagsInput{
				KeyId: key.KeyId,
			})
			if err != nil {
				innerErr = err
				return false
			}

			kmsKey.tags = tags.Tags
			resources = append(resources, kmsKey)
		}

		if lastPage {
			return false
		}

		return true
	}); err != nil {
		return nil, err
	}

	if innerErr != nil {
		return nil, innerErr
	}

	return resources, nil
}

type KMSKey struct {
	svc     *kms.KMS
	id      string
	state   string
	manager *string
	tags    []*kms.Tag
}

func (e *KMSKey) Filter() error {
	if e.state == "PendingDeletion" {
		return fmt.Errorf("is already in PendingDeletion state")
	}

	if e.manager != nil && *e.manager == kms.KeyManagerTypeAws {
		return fmt.Errorf("cannot delete AWS managed key")
	}

	return nil
}

func (e *KMSKey) Remove(_ context.Context) error {
	_, err := e.svc.ScheduleKeyDeletion(&kms.ScheduleKeyDeletionInput{
		KeyId:               &e.id,
		PendingWindowInDays: aws.Int64(7),
	})
	return err
}

func (e *KMSKey) String() string {
	return e.id
}

func (e *KMSKey) Properties() types.Properties {
	properties := types.NewProperties()
	properties.
		Set("ID", e.id)

	for _, tag := range e.tags {
		properties.SetTag(tag.TagKey, tag.TagValue)
	}

	return properties
}
