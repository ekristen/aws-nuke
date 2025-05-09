package resources

import (
	"context"
	"fmt"

	"github.com/gotidy/ptr"

	"github.com/aws/aws-sdk-go/service/route53resolver"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const Route53ResolverEndpointResource = "Route53ResolverEndpoint"

func init() {
	registry.Register(&registry.Registration{
		Name:     Route53ResolverEndpointResource,
		Scope:    nuke.Account,
		Resource: &Route53ResolverEndpoint{},
		Lister:   &Route53ResolverEndpointLister{},
	})
}

type Route53ResolverEndpointLister struct{}

// List produces the raw list of Route53 Resolver Endpoints to be nuked before filtering
func (l *Route53ResolverEndpointLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := route53resolver.New(opts.Session)

	params := &route53resolver.ListResolverEndpointsInput{}

	var resources []resource.Resource

	for {
		resp, err := svc.ListResolverEndpoints(params)

		if err != nil {
			return nil, err
		}

		for _, endpoint := range resp.ResolverEndpoints {
			resolverEndpoint := &Route53ResolverEndpoint{
				svc:  svc,
				id:   endpoint.Id,
				name: endpoint.Name,
			}

			resources = append(resources, resolverEndpoint)
		}

		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return resources, nil
}

// Route53ResolverEndpoint is the resource type for nuking
type Route53ResolverEndpoint struct {
	svc  *route53resolver.Route53Resolver
	id   *string
	name *string
}

// Remove implements Resource
func (r *Route53ResolverEndpoint) Remove(_ context.Context) error {
	_, err := r.svc.DeleteResolverEndpoint(
		&route53resolver.DeleteResolverEndpointInput{
			ResolverEndpointId: r.id,
		})

	if err != nil {
		return err
	}

	return nil
}

// Properties provides debugging output
func (r *Route53ResolverEndpoint) Properties() types.Properties {
	return types.NewProperties().
		Set("EndpointID", r.id).
		Set("Name", r.name)
}

// String implements Stringer
func (r *Route53ResolverEndpoint) String() string {
	return fmt.Sprintf("%s (%s)", ptr.ToString(r.id), ptr.ToString(r.name))
}
