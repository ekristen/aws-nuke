package resources

import (
	"context"
	"errors"

	"fmt"

	"github.com/gotidy/ptr"

	"github.com/aws/aws-sdk-go-v2/service/appsync"
	rtypes "github.com/aws/aws-sdk-go-v2/service/appsync/types"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const AppSyncAPIAssociationResource = "AppSyncAPIAssociation"

func init() {
	registry.Register(&registry.Registration{
		Name:     AppSyncAPIAssociationResource,
		Scope:    nuke.Account,
		Resource: &AppSyncAPIAssociation{},
		Lister:   &AppSyncAPIAssociationLister{},
	})
}

type AppSyncAPIAssociationLister struct{}

func (l *AppSyncAPIAssociationLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := appsync.NewFromConfig(*opts.Config)
	var resources []resource.Resource

	res, err := svc.ListDomainNames(ctx, &appsync.ListDomainNamesInput{})
	if err != nil {
		return nil, err
	}

	for _, p := range res.DomainNameConfigs {
		associationRes, err := svc.GetApiAssociation(ctx, &appsync.GetApiAssociationInput{
			DomainName: p.DomainName,
		})
		if err != nil {
			var notFound *rtypes.NotFoundException
			if !errors.As(err, &notFound) {
				return nil, err
			}
		} else {
			resources = append(resources, &AppSyncAPIAssociation{
				svc:        svc,
				DomainName: associationRes.ApiAssociation.DomainName,
				APIID:      associationRes.ApiAssociation.ApiId,
			})
		}
	}

	return resources, nil
}

type AppSyncAPIAssociation struct {
	svc        *appsync.Client
	DomainName *string
	APIID      *string
}

func (r *AppSyncAPIAssociation) Remove(ctx context.Context) error {
	_, err := r.svc.DisassociateApi(ctx, &appsync.DisassociateApiInput{
		DomainName: r.DomainName,
	})
	return err
}

func (r *AppSyncAPIAssociation) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *AppSyncAPIAssociation) String() string {
	return fmt.Sprintf("%s -> %s", ptr.ToString(r.DomainName), ptr.ToString(r.APIID))
}
