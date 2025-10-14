package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/lightsail" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const LightsailKeyPairResource = "LightsailKeyPair"

func init() {
	registry.Register(&registry.Registration{
		Name:     LightsailKeyPairResource,
		Scope:    nuke.Account,
		Resource: &LightsailKeyPair{},
		Lister:   &LightsailKeyPairLister{},
	})
}

type LightsailKeyPairLister struct{}

func (l *LightsailKeyPairLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := lightsail.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &lightsail.GetKeyPairsInput{}

	for {
		output, err := svc.GetKeyPairs(params)
		if err != nil {
			return nil, err
		}

		for _, keyPair := range output.KeyPairs {
			resources = append(resources, &LightsailKeyPair{
				svc:         svc,
				keyPairName: keyPair.Name,
			})
		}

		if output.NextPageToken == nil {
			break
		}

		params.PageToken = output.NextPageToken
	}

	return resources, nil
}

type LightsailKeyPair struct {
	svc         *lightsail.Lightsail
	keyPairName *string
}

func (f *LightsailKeyPair) Remove(_ context.Context) error {
	_, err := f.svc.DeleteKeyPair(&lightsail.DeleteKeyPairInput{
		KeyPairName: f.keyPairName,
	})

	return err
}

func (f *LightsailKeyPair) String() string {
	return *f.keyPairName
}
