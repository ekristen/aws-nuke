package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws/endpoints"     //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/lightsail" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

// TODO: implement region hints when we know certain things will never be in other regions

const LightsailDomainResource = "LightsailDomain"

func init() {
	registry.Register(&registry.Registration{
		Name:     LightsailDomainResource,
		Scope:    nuke.Account,
		Resource: &LightsailDomain{},
		Lister:   &LightsailDomainLister{},
	})
}

type LightsailDomainLister struct{}

func (l *LightsailDomainLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := lightsail.New(opts.Session)
	resources := make([]resource.Resource, 0)

	if opts.Session.Config.Region == nil || *opts.Session.Config.Region != endpoints.UsEast1RegionID {
		// LightsailDomain only supports us-east-1
		return resources, nil
	}

	params := &lightsail.GetDomainsInput{}

	for {
		output, err := svc.GetDomains(params)
		if err != nil {
			return nil, err
		}

		for _, domain := range output.Domains {
			resources = append(resources, &LightsailDomain{
				svc:        svc,
				domainName: domain.Name,
			})
		}

		if output.NextPageToken == nil {
			break
		}

		params.PageToken = output.NextPageToken
	}

	return resources, nil
}

type LightsailDomain struct {
	svc        *lightsail.Lightsail
	domainName *string
}

func (f *LightsailDomain) Remove(_ context.Context) error {
	_, err := f.svc.DeleteDomain(&lightsail.DeleteDomainInput{
		DomainName: f.domainName,
	})

	return err
}

func (f *LightsailDomain) String() string {
	return *f.domainName
}
