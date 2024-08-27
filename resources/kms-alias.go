package resources

import (
	"context"
	"fmt"
	"strings"

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
		Name:   KMSAliasResource,
		Scope:  nuke.Account,
		Lister: &KMSAliasLister{},
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
			resources = append(resources, &KMSAlias{
				svc:  svc,
				Name: alias.AliasName,
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
	svc  kmsiface.KMSAPI
	Name *string
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
