package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/lightsail"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const LightsailLoadBalancerResource = "LightsailLoadBalancer"

func init() {
	registry.Register(&registry.Registration{
		Name:   LightsailLoadBalancerResource,
		Scope:  nuke.Account,
		Lister: &LightsailLoadBalancerLister{},
	})
}

type LightsailLoadBalancerLister struct{}

func (l *LightsailLoadBalancerLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := lightsail.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &lightsail.GetLoadBalancersInput{}

	for {
		output, err := svc.GetLoadBalancers(params)
		if err != nil {
			return nil, err
		}

		for _, lb := range output.LoadBalancers {
			resources = append(resources, &LightsailLoadBalancer{
				svc:              svc,
				loadBalancerName: lb.Name,
			})
		}

		if output.NextPageToken == nil {
			break
		}

		params.PageToken = output.NextPageToken
	}

	return resources, nil
}

type LightsailLoadBalancer struct {
	svc              *lightsail.Lightsail
	loadBalancerName *string
}

func (f *LightsailLoadBalancer) Remove(_ context.Context) error {
	_, err := f.svc.DeleteLoadBalancer(&lightsail.DeleteLoadBalancerInput{
		LoadBalancerName: f.loadBalancerName,
	})

	return err
}

func (f *LightsailLoadBalancer) String() string {
	return *f.loadBalancerName
}
