package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/configservice" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const ConfigServiceConformancePackResource = "ConfigServiceConformancePack"

func init() {
	registry.Register(&registry.Registration{
		Name:     ConfigServiceConformancePackResource,
		Scope:    nuke.Account,
		Resource: &ConfigServiceConformancePack{},
		Lister:   &ConfigServiceConformancePackLister{},
	})
}

type ConfigServiceConformancePackLister struct{}

func (l *ConfigServiceConformancePackLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := configservice.New(opts.Session)
	var resources []resource.Resource

	var nextToken *string

	for {
		res, err := svc.DescribeConformancePacks(&configservice.DescribeConformancePacksInput{
			NextToken: nextToken,
		})
		if err != nil {
			return nil, err
		}

		for _, p := range res.ConformancePackDetails {
			resources = append(resources, &ConfigServiceConformancePack{
				svc:  svc,
				id:   p.ConformancePackId,
				name: p.ConformancePackName,
			})
		}

		if res.NextToken == nil {
			break
		}

		nextToken = res.NextToken
	}

	return resources, nil
}

type ConfigServiceConformancePack struct {
	svc  *configservice.ConfigService
	id   *string
	name *string
}

func (r *ConfigServiceConformancePack) Remove(_ context.Context) error {
	_, err := r.svc.DeleteConformancePack(&configservice.DeleteConformancePackInput{
		ConformancePackName: r.name,
	})
	return err
}

func (r *ConfigServiceConformancePack) Properties() types.Properties {
	props := types.NewProperties()
	props.Set("ID", r.id)
	props.Set("Name", r.name)
	return props
}

func (r *ConfigServiceConformancePack) String() string {
	return *r.id
}
