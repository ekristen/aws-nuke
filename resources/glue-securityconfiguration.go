package resources

import (
	"context"

	"github.com/gotidy/ptr"

	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/aws/aws-sdk-go/service/glue/glueiface"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const GlueSecurityConfigurationResource = "GlueSecurityConfiguration"

func init() {
	registry.Register(&registry.Registration{
		Name:   GlueSecurityConfigurationResource,
		Scope:  nuke.Account,
		Lister: &GlueSecurityConfigurationLister{},
	})
}

type GlueSecurityConfigurationLister struct {
	mockSvc glueiface.GlueAPI
}

func (l *GlueSecurityConfigurationLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	var resources []resource.Resource

	var svc glueiface.GlueAPI
	if l.mockSvc != nil {
		svc = l.mockSvc
	} else {
		svc = glue.New(opts.Session)
	}

	var nextToken *string
	for {
		params := &glue.GetSecurityConfigurationsInput{
			NextToken: nextToken,
		}

		res, err := svc.GetSecurityConfigurations(params)
		if err != nil {
			return nil, err
		}

		for _, p := range res.SecurityConfigurations {
			resources = append(resources, &GlueSecurityConfiguration{
				svc:  svc,
				name: p.Name,
			})
		}

		if res.NextToken == nil || ptr.ToString(res.NextToken) == "" {
			break
		}

		if len(res.SecurityConfigurations) == 0 {
			break
		}

		nextToken = res.NextToken
	}

	return resources, nil
}

type GlueSecurityConfiguration struct {
	svc  glueiface.GlueAPI
	name *string
}

func (r *GlueSecurityConfiguration) Remove(_ context.Context) error {
	_, err := r.svc.DeleteSecurityConfiguration(&glue.DeleteSecurityConfigurationInput{
		Name: r.name,
	})
	return err
}

func (r *GlueSecurityConfiguration) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("Name", r.name)
	return properties
}

func (r *GlueSecurityConfiguration) String() string {
	return *r.name
}
