package resources

import (
	"context"

	"github.com/gotidy/ptr"

	"github.com/aws/aws-sdk-go-v2/service/route53profiles"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const Route53ProfileResource = "Route53Profile"

func init() {
	registry.Register(&registry.Registration{
		Name:     Route53ProfileResource,
		Scope:    nuke.Account,
		Resource: &Route53Profile{},
		Lister:   &Route53ProfileLister{},
	})
}

type Route53ProfileLister struct{}

func (l *Route53ProfileLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := route53profiles.NewFromConfig(*opts.Config)
	var resources []resource.Resource

	params := &route53profiles.ListProfilesInput{
		MaxResults: ptr.Int32(100),
	}

	for {
		res, err := svc.ListProfiles(ctx, params)
		if err != nil {
			return nil, err
		}

		for _, p := range res.ProfileSummaries {
			var tags map[string]string
			tagsRes, err := svc.ListTagsForResource(ctx, &route53profiles.ListTagsForResourceInput{
				ResourceArn: p.Arn,
			})
			if err != nil {
				opts.Logger.Warnf("unable to fetch tags for profile: %s", ptr.ToString(p.Arn))
			} else {
				tags = tagsRes.Tags
			}

			resources = append(resources, &Route53Profile{
				svc:  svc,
				ID:   p.Id,
				Name: p.Name,
				Tags: tags,
			})
		}

		if res.NextToken == nil {
			break
		}

		params.NextToken = res.NextToken
	}

	return resources, nil
}

type Route53Profile struct {
	svc  *route53profiles.Client
	ID   *string
	Name *string
	Tags map[string]string
}

func (r *Route53Profile) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteProfile(ctx, &route53profiles.DeleteProfileInput{
		ProfileId: r.ID,
	})
	return err
}

func (r *Route53Profile) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *Route53Profile) String() string {
	return *r.Name
}
