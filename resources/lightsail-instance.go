package resources

import (
	"context"

	"github.com/gotidy/ptr"

	"github.com/aws/aws-sdk-go/service/lightsail" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	libsettings "github.com/ekristen/libnuke/pkg/settings"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const LightsailInstanceResource = "LightsailInstance"

func init() {
	registry.Register(&registry.Registration{
		Name:     LightsailInstanceResource,
		Scope:    nuke.Account,
		Resource: &LightsailInstance{},
		Lister:   &LightsailInstanceLister{},
		Settings: []string{
			"ForceDeleteAddOns",
		},
	})
}

type LightsailInstanceLister struct{}

func (l *LightsailInstanceLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := lightsail.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &lightsail.GetInstancesInput{}

	for {
		output, err := svc.GetInstances(params)
		if err != nil {
			return nil, err
		}

		for _, instance := range output.Instances {
			resources = append(resources, &LightsailInstance{
				svc:  svc,
				Name: instance.Name,
				Tags: instance.Tags,
			})
		}

		if output.NextPageToken == nil {
			break
		}

		params.PageToken = output.NextPageToken
	}

	return resources, nil
}

type LightsailInstance struct {
	svc  *lightsail.Lightsail
	Name *string `description:"The name of the instance."`
	Tags []*lightsail.Tag

	settings *libsettings.Setting
}

func (r *LightsailInstance) Settings(setting *libsettings.Setting) {
	r.settings = setting
}

func (r *LightsailInstance) Remove(_ context.Context) error {
	_, err := r.svc.DeleteInstance(&lightsail.DeleteInstanceInput{
		InstanceName:      r.Name,
		ForceDeleteAddOns: ptr.Bool(r.settings.GetBool("ForceDeleteAddOns")),
	})

	return err
}

func (r *LightsailInstance) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *LightsailInstance) String() string {
	return *r.Name
}
