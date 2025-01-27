package resources

import (
	"context"
	"fmt"

	"github.com/gotidy/ptr"

	"github.com/aws/aws-sdk-go/service/route53"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const Route53HostedZoneResource = "Route53HostedZone"

func init() {
	registry.Register(&registry.Registration{
		Name:     Route53HostedZoneResource,
		Scope:    nuke.Account,
		Resource: &Route53HostedZone{},
		Lister:   &Route53HostedZoneLister{},
	})
}

type Route53HostedZoneLister struct{}

func (l *Route53HostedZoneLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := route53.New(opts.Session)

	var hostedZones []*route53.HostedZone
	params := &route53.ListHostedZonesInput{}

	for {
		resp, err := svc.ListHostedZones(params)
		if err != nil {
			return nil, err
		}

		hostedZones = append(hostedZones, resp.HostedZones...)

		params.Marker = resp.NextMarker
		if ptr.ToString(params.Marker) == "" {
			break
		}
	}

	resources := make([]resource.Resource, 0)
	for _, hz := range hostedZones {
		tags, err := svc.ListTagsForResource(&route53.ListTagsForResourceInput{
			ResourceId:   hz.Id,
			ResourceType: ptr.String("hostedzone"),
		})

		if err != nil {
			return nil, err
		}

		resources = append(resources, &Route53HostedZone{
			svc:  svc,
			ID:   hz.Id,
			Name: hz.Name,
			Tags: tags.ResourceTagSet.Tags,
		})
	}
	return resources, nil
}

type Route53HostedZone struct {
	svc  *route53.Route53
	ID   *string
	Name *string
	Tags []*route53.Tag
}

func (r *Route53HostedZone) Remove(_ context.Context) error {
	params := &route53.DeleteHostedZoneInput{
		Id: r.ID,
	}

	_, err := r.svc.DeleteHostedZone(params)
	if err != nil {
		return err
	}

	return nil
}

func (r *Route53HostedZone) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *Route53HostedZone) String() string {
	// TODO(v4): change the stringer format
	return fmt.Sprintf("%s (%s)", *r.ID, *r.Name)
}
