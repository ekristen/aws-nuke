package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/lightsail"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const LightsailStaticIPResource = "LightsailStaticIP"

func init() {
	registry.Register(&registry.Registration{
		Name:     LightsailStaticIPResource,
		Scope:    nuke.Account,
		Resource: &LightsailStaticIP{},
		Lister:   &LightsailStaticIPLister{},
	})
}

type LightsailStaticIPLister struct{}

func (l *LightsailStaticIPLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := lightsail.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &lightsail.GetStaticIpsInput{}

	for {
		output, err := svc.GetStaticIps(params)
		if err != nil {
			return nil, err
		}

		for _, staticIP := range output.StaticIps {
			resources = append(resources, &LightsailStaticIP{
				svc:          svc,
				staticIPName: staticIP.Name,
			})
		}

		if output.NextPageToken == nil {
			break
		}

		params.PageToken = output.NextPageToken
	}

	return resources, nil
}

type LightsailStaticIP struct {
	svc          *lightsail.Lightsail
	staticIPName *string
}

func (f *LightsailStaticIP) Remove(_ context.Context) error {
	_, err := f.svc.ReleaseStaticIp(&lightsail.ReleaseStaticIpInput{
		StaticIpName: f.staticIPName,
	})

	return err
}

func (f *LightsailStaticIP) String() string {
	return *f.staticIPName
}
