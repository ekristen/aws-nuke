package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloud9"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

type Cloud9Environment struct {
	svc           *cloud9.Cloud9
	environmentID *string
}

const Cloud9EnvironmentResource = "Cloud9Environment"

func init() {
	registry.Register(&registry.Registration{
		Name:   Cloud9EnvironmentResource,
		Scope:  nuke.Account,
		Lister: &Cloud9EnvironmentLister{},
	})
}

type Cloud9EnvironmentLister struct{}

func (l *Cloud9EnvironmentLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := cloud9.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &cloud9.ListEnvironmentsInput{
		MaxResults: aws.Int64(25),
	}

	for {
		resp, err := svc.ListEnvironments(params)
		if err != nil {
			return nil, err
		}

		for _, environmentID := range resp.EnvironmentIds {
			resources = append(resources, &Cloud9Environment{
				svc:           svc,
				environmentID: environmentID,
			})
		}

		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return resources, nil
}

func (f *Cloud9Environment) Remove(_ context.Context) error {
	_, err := f.svc.DeleteEnvironment(&cloud9.DeleteEnvironmentInput{
		EnvironmentId: f.environmentID,
	})

	return err
}

func (f *Cloud9Environment) String() string {
	return *f.environmentID
}
