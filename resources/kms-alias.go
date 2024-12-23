package resources

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/kms/kmsiface"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const KMSAliasResource = "KMSAlias"

func init() {
	registry.Register(&registry.Registration{
		Name:     KMSAliasResource,
		Scope:    nuke.Account,
		Resource: &KMSAlias{},
		Lister:   &KMSAliasLister{},
	})
}

type KMSAliasLister struct {
	mockSvc kmsiface.KMSAPI
}

func (l *KMSAliasLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	resources := make([]resource.Resource, 0)

	var svc kmsiface.KMSAPI
	if l.mockSvc != nil {
		svc = l.mockSvc
	} else {
		svc = kms.New(opts.Session)
	}

	err := svc.ListAliasesPages(nil, func(page *kms.ListAliasesOutput, lastPage bool) bool {
		for _, alias := range page.Aliases {
			var tags []*kms.Tag
			if alias.TargetKeyId != nil {
				keyTags, err := svc.ListResourceTags(&kms.ListResourceTagsInput{
					KeyId: alias.TargetKeyId,
				})
				if err != nil {
					opts.Logger.WithError(err).Error("failed to list tags for key")
				}
				tags = keyTags.Tags
			}

			resources = append(resources, &KMSAlias{
				svc:  svc,
				Name: alias.AliasName,
				Tags: tags,
			})
		}
		return true
	})
	if err != nil {
		return nil, err
	}

	return resources, nil
}

type KMSAlias struct {
	svc          kmsiface.KMSAPI
	Name         *string    `description:"The name of the KMS alias"`
	CreationDate *time.Time `description:"The creation date of the KMS alias"`
	TargetKeyID  *string    `description:"The KMS Key ID that the alias points to"`
	Tags         []*kms.Tag `property:"tagPrefix=key:tag"`
}

func (r *KMSAlias) Filter() error {
	if strings.HasPrefix(*r.Name, "alias/aws/") {
		return fmt.Errorf("cannot delete AWS alias")
	}
	return nil
}

func (r *KMSAlias) Remove(_ context.Context) error {
	_, err := r.svc.DeleteAlias(&kms.DeleteAliasInput{
		AliasName: r.Name,
	})
	return err
}

func (r *KMSAlias) String() string {
	return *r.Name
}

func (r *KMSAlias) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}
