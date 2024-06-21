package resources

import (
	"context"

	"fmt"
	"strings"

	"github.com/gotidy/ptr"

	"github.com/aws/aws-sdk-go/service/kms"

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

type KMSAliasLister struct{}

func (l *KMSAliasLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := kms.New(opts.Session)

	resp, err := svc.ListAliases(nil)
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, 0)
	for _, alias := range resp.Aliases {
		resources = append(resources, &KMSAlias{
			svc:  svc,
			name: ptr.ToString(alias.AliasName),
		})
	}

	return resources, nil
}

type KMSAlias struct {
	svc  *kms.KMS
	name string
}

func (e *KMSAlias) Filter() error {
	if strings.HasPrefix(e.name, "alias/aws/") {
		return fmt.Errorf("cannot delete AWS alias")
	}
	return nil
}

func (e *KMSAlias) Remove(_ context.Context) error {
	_, err := e.svc.DeleteAlias(&kms.DeleteAliasInput{
		AliasName: &e.name,
	})
	return err
}

func (e *KMSAlias) String() string {
	return e.name
}

func (e *KMSAlias) Properties() types.Properties {
	properties := types.NewProperties()
	properties.
		Set("Name", e.name)

	return properties
}
