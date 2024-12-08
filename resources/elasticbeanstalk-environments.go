package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elasticbeanstalk"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const ElasticBeanstalkEnvironmentResource = "ElasticBeanstalkEnvironment"

func init() {
	registry.Register(&registry.Registration{
		Name:     ElasticBeanstalkEnvironmentResource,
		Scope:    nuke.Account,
		Resource: &ElasticBeanstalkEnvironment{},
		Lister:   &ElasticBeanstalkEnvironmentLister{},
	})
}

type ElasticBeanstalkEnvironmentLister struct{}

func (l *ElasticBeanstalkEnvironmentLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := elasticbeanstalk.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &elasticbeanstalk.DescribeEnvironmentsInput{
		MaxRecords:     aws.Int64(100),
		IncludeDeleted: aws.Bool(false),
	}

	for {
		output, err := svc.DescribeEnvironments(params)
		if err != nil {
			return nil, err
		}

		for _, environment := range output.Environments {
			resources = append(resources, &ElasticBeanstalkEnvironment{
				svc:  svc,
				ID:   environment.EnvironmentId,
				name: environment.EnvironmentName,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type ElasticBeanstalkEnvironment struct {
	svc  *elasticbeanstalk.ElasticBeanstalk
	ID   *string
	name *string
}

func (f *ElasticBeanstalkEnvironment) Remove(_ context.Context) error {
	_, err := f.svc.TerminateEnvironment(&elasticbeanstalk.TerminateEnvironmentInput{
		EnvironmentId:      f.ID,
		ForceTerminate:     aws.Bool(true),
		TerminateResources: aws.Bool(true),
	})

	return err
}

func (f *ElasticBeanstalkEnvironment) Properties() types.Properties {
	return types.NewProperties().
		Set("Name", f.name)
}

func (f *ElasticBeanstalkEnvironment) String() string {
	return *f.ID
}
