package resources

import (
	"context"
	"strings"
	"time"

	"github.com/gotidy/ptr"

	"github.com/aws/aws-sdk-go-v2/service/s3control"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const S3AccessGrantsGrantResource = "S3AccessGrantsGrant"

func init() {
	registry.Register(&registry.Registration{
		Name:     S3AccessGrantsGrantResource,
		Scope:    nuke.Account,
		Resource: &S3AccessGrantsGrant{},
		Lister:   &S3AccessGrantsGrantLister{},
	})
}

type S3AccessGrantsGrantLister struct{}

func (l *S3AccessGrantsGrantLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := s3control.NewFromConfig(*opts.Config)
	var resources []resource.Resource

	res, err := svc.ListAccessGrants(ctx, &s3control.ListAccessGrantsInput{
		AccountId: opts.AccountID,
	})
	if err != nil {
		if strings.Contains(err.Error(), "AccessGrantsInstanceNotExistsError") {
			return resources, nil
		} else {
			return nil, err
		}
	}

	for _, p := range res.AccessGrantsList {
		resources = append(resources, &S3AccessGrantsGrant{
			svc:         svc,
			accountID:   opts.AccountID,
			ID:          p.AccessGrantId,
			GrantScope:  p.GrantScope,
			GranteeType: ptr.String(string(p.Grantee.GranteeType)),
			GranteeID:   p.Grantee.GranteeIdentifier,
			CreatedAt:   p.CreatedAt,
		})
	}

	return resources, nil
}

type S3AccessGrantsGrant struct {
	svc         *s3control.Client
	accountID   *string
	ID          *string    `description:"The ID of the access grant."`
	GrantScope  *string    `description:"The scope of the access grant."`
	GranteeType *string    `description:"The type of the grantee, (e.g. IAM)."`
	GranteeID   *string    `description:"The ARN of the grantee."`
	CreatedAt   *time.Time `description:"The date and time the access grant was created."`
}

func (r *S3AccessGrantsGrant) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteAccessGrant(ctx, &s3control.DeleteAccessGrantInput{
		AccountId:     r.accountID,
		AccessGrantId: r.ID,
	})
	return err
}

func (r *S3AccessGrantsGrant) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *S3AccessGrantsGrant) String() string {
	return *r.ID
}
