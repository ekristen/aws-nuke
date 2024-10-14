package resources

import (
	"context"
	"errors"
	"fmt"

	"github.com/gotidy/ptr"
	"github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/kms/kmsiface"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
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

type KMSKeyLister struct {
	mockSvc kmsiface.KMSAPI
}

func (l *KMSKeyLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	resources := make([]resource.Resource, 0)

	var svc kmsiface.KMSAPI
	if l.mockSvc != nil {
		svc = l.mockSvc
	} else {
		svc = kms.New(opts.Session)
	}

	inaccessibleKeys := false

	if err := svc.ListKeysPages(nil, func(keysOut *kms.ListKeysOutput, lastPage bool) bool {
		for _, key := range keysOut.Keys {
			resp, err := svc.DescribeKey(&kms.DescribeKeyInput{
				KeyId: key.KeyId,
			})
			if err != nil {
				var awsError awserr.Error
				if errors.As(err, &awsError) {
					if awsError.Code() == "AccessDeniedException" {
						inaccessibleKeys = true
						logrus.WithField("arn", key.KeyArn).WithError(err).Debug("unable to describe key")
						continue
					}
				}

				logrus.WithError(err).Error("unable to describe key")
				continue
			}

			kmsKey := &KMSKey{
				svc:     svc,
				ID:      resp.KeyMetadata.KeyId,
				State:   resp.KeyMetadata.KeyState,
				Manager: resp.KeyMetadata.KeyManager,
			}

			// Note: we check for customer managed keys here because we can't list tags for AWS managed keys
			// This way AWS managed keys still show up but get filtered out by the Filter method
			if ptr.ToString(resp.KeyMetadata.KeyManager) == kms.KeyManagerTypeCustomer {
				tags, err := svc.ListResourceTags(&kms.ListResourceTagsInput{
					KeyId: key.KeyId,
				})
				if err != nil {
					var awsError awserr.Error
					if errors.As(err, &awsError) {
						if awsError.Code() == "AccessDeniedException" {
							inaccessibleKeys = true
							logrus.WithError(err).Debug("unable to list tags")
							continue
						} else {
							logrus.WithError(err).Error("unable to list tags")
						}
					}
				} else {
					kmsKey.Tags = tags.Tags
				}
			}

			keyAliases, err := svc.ListAliases(&kms.ListAliasesInput{
				KeyId: key.KeyId,
			})
			if err != nil {
				logrus.WithError(err).Error("unable to list aliases")
			}

			if len(keyAliases.Aliases) > 0 {
				kmsKey.Alias = keyAliases.Aliases[0].AliasName
			}

			resources = append(resources, kmsKey)
		}

		return !lastPage
	}); err != nil {
		return nil, err
	}

	if inaccessibleKeys {
		logrus.Warn("one or more KMS keys were inaccessible, debug logging will contain more information")
	}

	return resources, nil
}

type KMSKey struct {
	svc     kmsiface.KMSAPI
	ID      *string
	State   *string
	Manager *string
	Alias   *string
	Tags    []*kms.Tag
}

func (r *KMSKey) Filter() error {
	if ptr.ToString(r.State) == kms.KeyStatePendingDeletion {
		return fmt.Errorf("is already in PendingDeletion state")
	}

	if ptr.ToString(r.Manager) == kms.KeyManagerTypeAws {
		return fmt.Errorf("cannot delete AWS managed key")
	}

	return nil
}

func (r *KMSKey) Remove(_ context.Context) error {
	_, err := r.svc.ScheduleKeyDeletion(&kms.ScheduleKeyDeletionInput{
		KeyId:               r.ID,
		PendingWindowInDays: aws.Int64(7),
	})
	return err
}

func (r *KMSKey) String() string {
	return *r.ID
}

func (r *KMSKey) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}
