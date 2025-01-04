package resources

import (
	"context"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3control"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const S3AccessGrantsLocationResource = "S3AccessGrantsLocation"

func init() {
	registry.Register(&registry.Registration{
		Name:     S3AccessGrantsLocationResource,
		Scope:    nuke.Account,
		Resource: &S3AccessGrantsLocation{},
		Lister:   &S3AccessGrantsLocationLister{},
	})
}

type S3AccessGrantsLocationLister struct{}

func (l *S3AccessGrantsLocationLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := s3control.NewFromConfig(*opts.Config)
	var resources []resource.Resource

	res, err := svc.ListAccessGrantsLocations(ctx, &s3control.ListAccessGrantsLocationsInput{
		AccountId: opts.AccountID,
	})
	if err != nil {
		if strings.Contains(err.Error(), "AccessGrantsInstanceNotExistsError") {
			return resources, nil
		} else {
			return nil, err
		}
	}

	for _, entity := range res.AccessGrantsLocationsList {
		resources = append(resources, &S3AccessGrantsLocation{
			svc:           svc,
			accountID:     opts.AccountID,
			ID:            entity.AccessGrantsLocationId,
			LocationScope: entity.LocationScope,
			CreatedAt:     entity.CreatedAt,
		})
	}

	return resources, nil
}

type S3AccessGrantsLocation struct {
	svc           *s3control.Client
	accountID     *string
	ID            *string    `description:"The ID of the access grants location."`
	LocationScope *string    `description:"The scope of the access grants location."`
	CreatedAt     *time.Time `description:"The time the access grants location was created."`
}

func (r *S3AccessGrantsLocation) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteAccessGrantsLocation(ctx, &s3control.DeleteAccessGrantsLocationInput{
		AccessGrantsLocationId: r.ID,
		AccountId:              r.accountID,
	})
	return err
}

func (r *S3AccessGrantsLocation) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *S3AccessGrantsLocation) String() string {
	return *r.ID
}
