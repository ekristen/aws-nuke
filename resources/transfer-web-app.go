package resources

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/transfer"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const TransferWebAppResource = "TransferWebApp"

func init() {
	registry.Register(&registry.Registration{
		Name:     TransferWebAppResource,
		Scope:    nuke.Account,
		Resource: &TransferWebApp{},
		Lister:   &TransferWebAppLister{},
	})
}

type TransferWebAppLister struct{}

func (l *TransferWebAppLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := transfer.NewFromConfig(*opts.Config)
	var resources []resource.Resource

	res, err := svc.ListWebApps(ctx, &transfer.ListWebAppsInput{})
	if err != nil {
		return nil, err
	}

	for _, entry := range res.WebApps {
		resources = append(resources, &TransferWebApp{
			svc: svc,
			ID:  entry.WebAppId,
		})
	}

	return resources, nil
}

type TransferWebApp struct {
	svc *transfer.Client
	ID  *string
}

func (r *TransferWebApp) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteWebApp(ctx, &transfer.DeleteWebAppInput{
		WebAppId: r.ID,
	})
	return err
}

func (r *TransferWebApp) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *TransferWebApp) String() string {
	return *r.ID
}
