package resources

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/gotidy/ptr"
	"github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws"         //nolint:staticcheck
	"github.com/aws/aws-sdk-go/aws/awserr"  //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/kms" //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/kms/kmsiface"

	liberrors "github.com/ekristen/libnuke/pkg/errors"
	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	libsettings "github.com/ekristen/libnuke/pkg/settings"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const KMSKeyResource = "KMSKey"

func init() {
	registry.Register(&registry.Registration{
		Name:     KMSKeyResource,
		Scope:    nuke.Account,
		Resource: &KMSKey{},
		Lister:   &KMSKeyLister{},
		DependsOn: []string{
			KMSAliasResource,
		},
		Settings: []string{
			"DeleteInUseKeys",
		},
	})
}

// awsServicePrincipalSuffix is the suffix of an AWS service principal, used to tell grants held by a service apart
// from grants held by an IAM principal.
const awsServicePrincipalSuffix = ".amazonaws.com"

// serviceGrantees returns the distinct AWS service principals holding grants on the key.
//
// Services that encrypt data at rest with a customer managed key (RDS, EBS, EFS, DynamoDB, and others) create a grant
// on the key when the encrypted resource is created, and retire it when the resource is deleted. An outstanding
// service grant therefore means a live resource still depends on this key, and scheduling the key for deletion would
// leave that resource unusable.
//
// Grants held by IAM principals are deliberately ignored. They are not tied to the lifecycle of any resource, so
// counting them would block the key from ever being deleted.
func serviceGrantees(svc kmsiface.KMSAPI, keyID *string) ([]string, error) {
	seen := make(map[string]struct{})

	if err := svc.ListGrantsPages(&kms.ListGrantsInput{
		KeyId: keyID,
	}, func(page *kms.ListGrantsResponse, lastPage bool) bool {
		for _, grant := range page.Grants {
			principal := ptr.ToString(grant.GranteePrincipal)
			if strings.HasSuffix(principal, awsServicePrincipalSuffix) {
				seen[principal] = struct{}{}
			}
		}

		return !lastPage
	}); err != nil {
		return nil, err
	}

	grantees := make([]string, 0, len(seen))
	for principal := range seen {
		grantees = append(grantees, principal)
	}
	sort.Strings(grantees)

	return grantees, nil
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
							logrus.WithError(err).Debug("unable to list tags - inaccessible key")
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
			} else if len(keyAliases.Aliases) > 0 {
				kmsKey.Alias = keyAliases.Aliases[0].AliasName
			}

			// InUse is populated best effort so that it is available for filtering and for dry-run visibility. The
			// authoritative check happens again at removal time, because grants are retired while the run is in
			// progress as the resources holding them are deleted.
			grantees, err := serviceGrantees(svc, key.KeyId)
			if err != nil {
				logrus.WithField("key", ptr.ToString(key.KeyId)).WithError(err).Error("unable to list grants")
			} else {
				kmsKey.InUse = ptr.Bool(len(grantees) > 0)
				kmsKey.InUseBy = ptr.String(strings.Join(grantees, ","))
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
	svc      kmsiface.KMSAPI
	settings *libsettings.Setting
	ID       *string
	State    *string
	Manager  *string
	Alias    *string
	InUse    *bool   `description:"Whether an AWS service still holds a grant on the key, meaning a live resource depends on it."`
	InUseBy  *string `description:"Comma separated list of the AWS service principals holding grants on the key."`
	Tags     []*kms.Tag
}

func (r *KMSKey) Settings(setting *libsettings.Setting) {
	r.settings = setting
}

// deleteInUseKeys reports whether the operator has opted into deleting keys that AWS services still hold grants on.
func (r *KMSKey) deleteInUseKeys() bool {
	return r.settings != nil && r.settings.GetBool("DeleteInUseKeys")
}

// Filter is only for state that cannot change over the course of a run. Grant state is deliberately not checked here:
// it changes as the resources holding the grants are deleted, and a filtered resource is never reconsidered. See
// Remove for the in-use check.
func (r *KMSKey) Filter() error {
	if state := ptr.ToString(r.State); state == kms.KeyStatePendingDeletion || state == kms.KeyStatePendingReplicaDeletion {
		return fmt.Errorf("is already in %v state", state)
	}

	if ptr.ToString(r.Manager) == kms.KeyManagerTypeAws {
		return fmt.Errorf("cannot delete AWS managed key")
	}

	return nil
}

func (r *KMSKey) Remove(_ context.Context) error {
	if !r.deleteInUseKeys() {
		grantees, err := serviceGrantees(r.svc, r.ID)
		switch {
		case err != nil:
			// If grant state cannot be determined we let the deletion proceed rather than holding the key forever.
			// This preserves the behavior that existed before the in-use check was added.
			logrus.WithField("key", ptr.ToString(r.ID)).WithError(err).
				Warn("unable to list grants, proceeding with deletion")
		case len(grantees) > 0:
			// Held rather than filtered, so that the key is retried on each pass of the queue and is removed within
			// this run once the resources holding the grants have been deleted.
			return liberrors.ErrHoldResource(fmt.Sprintf("in use by %s", strings.Join(grantees, ", ")))
		}
	}

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
