package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/emr"

	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const EMRSecurityConfigurationResource = "EMRSecurityConfiguration"

func init() {
	resource.Register(resource.Registration{
		Name:   EMRSecurityConfigurationResource,
		Scope:  nuke.Account,
		Lister: &EMRSecurityConfigurationLister{},
	})
}

type EMRSecurityConfigurationLister struct{}

func (l *EMRSecurityConfigurationLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := emr.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &emr.ListSecurityConfigurationsInput{}

	for {
		resp, err := svc.ListSecurityConfigurations(params)
		if err != nil {
			return nil, err
		}

		for _, securityConfiguration := range resp.SecurityConfigurations {
			resources = append(resources, &EMRSecurityConfiguration{
				svc:  svc,
				name: securityConfiguration.Name,
			})
		}

		if resp.Marker == nil {
			break
		}

		params.Marker = resp.Marker
	}

	return resources, nil
}

type EMRSecurityConfiguration struct {
	svc  *emr.EMR
	name *string
}

func (f *EMRSecurityConfiguration) Remove(_ context.Context) error {
	// Note: Call names are inconsistent in the SDK
	_, err := f.svc.DeleteSecurityConfiguration(&emr.DeleteSecurityConfigurationInput{
		Name: f.name,
	})

	return err
}

func (f *EMRSecurityConfiguration) String() string {
	return *f.name
}
