package resources

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gotidy/ptr"

	"github.com/aws/aws-sdk-go-v2/service/route53profiles"
	rtypes "github.com/aws/aws-sdk-go-v2/service/route53profiles/types"

	liberror "github.com/ekristen/libnuke/pkg/errors"
	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const Route53ProfileAssociationResource = "Route53ProfileAssociation"

func init() {
	registry.Register(&registry.Registration{
		Name:     Route53ProfileAssociationResource,
		Scope:    nuke.Account,
		Resource: &Route53ProfileAssociation{},
		Lister:   &Route53ProfileAssociationLister{},
	})
}

type Route53ProfileAssociationLister struct{}

func (l *Route53ProfileAssociationLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := route53profiles.NewFromConfig(*opts.Config)
	var resources []resource.Resource

	params := &route53profiles.ListProfileAssociationsInput{
		MaxResults: ptr.Int32(100),
	}

	for {
		res, err := svc.ListProfileAssociations(ctx, params)
		if err != nil {
			return nil, err
		}

		for _, p := range res.ProfileAssociations {
			resources = append(resources, &Route53ProfileAssociation{
				svc:              svc,
				ID:               p.Id,
				Name:             p.Name,
				OwnerID:          p.OwnerId,
				ProfileID:        p.ProfileId,
				ResourceID:       p.ResourceId,
				Status:           ptr.String(fmt.Sprintf("%s", p.Status)),
				CreationTime:     p.CreationTime,
				ModificationTime: p.ModificationTime,
			})
		}

		if res.NextToken == nil {
			break
		}

		params.NextToken = res.NextToken
	}

	return resources, nil
}

type Route53ProfileAssociation struct {
	svc              *route53profiles.Client
	ID               *string
	Name             *string
	OwnerID          *string
	ProfileID        *string
	ResourceID       *string
	Status           *string
	CreationTime     *time.Time
	ModificationTime *time.Time
}

func (r *Route53ProfileAssociation) Filter() error {
	if ptr.ToString(r.Status) == string(rtypes.ProfileStatusCreating) {
		return errors.New("cannot delete profile association in CREATING state")
	}

	return nil
}

func (r *Route53ProfileAssociation) Remove(ctx context.Context) error {
	// Note: if somehow the status is already deleting we do not want to try to delete again. However, this is the
	// first resource to take advantage of the HandleWait method since deletion is not immediate and the disassociation
	// is not immediate.
	if ptr.ToString(r.Status) == string(rtypes.ProfileStatusDeleting) {
		return nil
	}

	// Note: disassociation is not immediate.
	_, err := r.svc.DisassociateProfile(ctx, &route53profiles.DisassociateProfileInput{
		ProfileId:  r.ProfileID,
		ResourceId: r.ResourceID,
	})
	return err
}

func (r *Route53ProfileAssociation) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *Route53ProfileAssociation) String() string {
	return *r.Name
}

func (r *Route53ProfileAssociation) HandleWait(ctx context.Context) error {
	var notFound *rtypes.ResourceNotFoundException
	p, err := r.svc.GetProfileAssociation(ctx, &route53profiles.GetProfileAssociationInput{
		ProfileAssociationId: r.ID,
	})
	if err != nil {
		if errors.As(err, &notFound) {
			return nil
		}

		return err
	}

	currentStatus := fmt.Sprintf("%s", p.ProfileAssociation.Status)

	r.Status = ptr.String(currentStatus)

	if currentStatus == string(rtypes.ProfileStatusDeleting) {
		return liberror.ErrWaitResource("waiting for operation to complete")
	}

	return nil
}
