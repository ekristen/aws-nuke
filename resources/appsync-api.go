package resources

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/appsync"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const AppSyncAPIResource = "AppSyncAPI"

func init() {
	registry.Register(&registry.Registration{
		Name:     AppSyncAPIResource,
		Scope:    nuke.Account,
		Resource: &AppSyncAPI{},
		Lister:   &AppSyncAPILister{},
	})
}

type AppSyncAPILister struct{}

func (l *AppSyncAPILister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := appsync.NewFromConfig(*opts.Config)
	var resources []resource.Resource

	params := &appsync.ListApisInput{}

	for {
		resp, err := svc.ListApis(ctx, params)
		if err != nil {
			return nil, err
		}

		for _, p := range resp.Apis {
			resources = append(resources, &AppSyncAPI{
				svc:  svc,
				ID:   p.ApiId,
				Tags: p.Tags,
			})
		}

		if resp.NextToken == nil {
			break
		}
		params.NextToken = resp.NextToken
	}

	return resources, nil
}

type AppSyncAPI struct {
	svc  *appsync.Client
	ID   *string
	Tags map[string]string
}

func (r *AppSyncAPI) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteApi(ctx, &appsync.DeleteApiInput{
		ApiId: r.ID,
	})
	return err
}

func (r *AppSyncAPI) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *AppSyncAPI) String() string {
	return *r.ID
}
