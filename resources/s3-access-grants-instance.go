package resources

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3control"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const S3AccessGrantsInstanceResource = "S3AccessGrantsInstance"

func init() {
	registry.Register(&registry.Registration{
		Name:     S3AccessGrantsInstanceResource,
		Scope:    nuke.Account,
		Resource: &S3AccessGrantsInstance{},
		Lister:   &S3AccessGrantsInstanceLister{},
	})
}

type S3AccessGrantsInstanceLister struct{}

func (l *S3AccessGrantsInstanceLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := s3control.NewFromConfig(*opts.Config)
	var resources []resource.Resource

	res, err := svc.ListAccessGrantsInstances(ctx, &s3control.ListAccessGrantsInstancesInput{
		AccountId: opts.AccountID,
	})
	if err != nil {
		return nil, err
	}

	for _, entity := range res.AccessGrantsInstancesList {
		resources = append(resources, &S3AccessGrantsInstance{
			svc:       svc,
			accountID: opts.AccountID,
			ID:        entity.AccessGrantsInstanceId,
			CreatedAt: entity.CreatedAt,
		})
	}

	return resources, nil
}

type S3AccessGrantsInstance struct {
	svc       *s3control.Client
	accountID *string
	ID        *string    `description:"The ID of the access grants instance."`
	CreatedAt *time.Time `description:"The time the access grants instance was created."`
}

func (r *S3AccessGrantsInstance) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteAccessGrantsInstance(ctx, &s3control.DeleteAccessGrantsInstanceInput{
		AccountId: r.accountID,
	})
	return err
}

func (r *S3AccessGrantsInstance) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *S3AccessGrantsInstance) String() string {
	return *r.ID
}
