package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/elasticbeanstalk"

	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const ElasticBeanstalkApplicationResource = "ElasticBeanstalkApplication"

func init() {
	resource.Register(&resource.Registration{
		Name:   ElasticBeanstalkApplicationResource,
		Scope:  nuke.Account,
		Lister: &ElasticBeanstalkApplicationLister{},
	})
}

type ElasticBeanstalkApplicationLister struct{}

func (l *ElasticBeanstalkApplicationLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := elasticbeanstalk.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &elasticbeanstalk.DescribeApplicationsInput{}

	output, err := svc.DescribeApplications(params)
	if err != nil {
		return nil, err
	}

	for _, application := range output.Applications {
		resources = append(resources, &ElasticBeanstalkApplication{
			svc:  svc,
			name: application.ApplicationName,
		})
	}

	return resources, nil
}

type ElasticBeanstalkApplication struct {
	svc  *elasticbeanstalk.ElasticBeanstalk
	name *string
}

func (f *ElasticBeanstalkApplication) Remove(_ context.Context) error {
	_, err := f.svc.DeleteApplication(&elasticbeanstalk.DeleteApplicationInput{
		ApplicationName: f.name,
	})

	return err
}

func (f *ElasticBeanstalkApplication) String() string {
	return *f.name
}
