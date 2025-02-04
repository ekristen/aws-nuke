package resources

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/appsync"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const AppSyncDomainNameResource = "AppSyncDomainName"

func init() {
	registry.Register(&registry.Registration{
		Name:     AppSyncDomainNameResource,
		Scope:    nuke.Account,
		Resource: &AppSyncDomainName{},
		Lister:   &AppSyncDomainNameLister{},
	})
}

type AppSyncDomainNameLister struct{}

func (l *AppSyncDomainNameLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := appsync.NewFromConfig(*opts.Config)
	var resources []resource.Resource

	res, err := svc.ListDomainNames(ctx, &appsync.ListDomainNamesInput{})
	if err != nil {
		return nil, err
	}

	for _, p := range res.DomainNameConfigs {
		resources = append(resources, &AppSyncDomainName{
			svc:        svc,
			DomainName: p.DomainName,
		})
	}

	return resources, nil
}

type AppSyncDomainName struct {
	svc        *appsync.Client
	DomainName *string
}

func (r *AppSyncDomainName) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteDomainName(ctx, &appsync.DeleteDomainNameInput{
		DomainName: r.DomainName,
	})
	return err
}

func (r *AppSyncDomainName) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *AppSyncDomainName) String() string {
	return *r.DomainName
}
