package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicediscovery"

	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const ServiceDiscoveryServiceResource = "ServiceDiscoveryService"

func init() {
	resource.Register(resource.Registration{
		Name:   ServiceDiscoveryServiceResource,
		Scope:  nuke.Account,
		Lister: &ServiceDiscoveryServiceLister{},
	})
}

type ServiceDiscoveryServiceLister struct{}

func (l *ServiceDiscoveryServiceLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := servicediscovery.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &servicediscovery.ListServicesInput{
		MaxResults: aws.Int64(100),
	}

	// Collect all services, using separate for loop
	// due to multi-service pagination issues
	for {
		output, err := svc.ListServices(params)
		if err != nil {
			return nil, err
		}

		for _, service := range output.Services {
			resources = append(resources, &ServiceDiscoveryService{
				svc: svc,
				ID:  service.Id,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type ServiceDiscoveryService struct {
	svc *servicediscovery.ServiceDiscovery
	ID  *string
}

func (f *ServiceDiscoveryService) Remove(_ context.Context) error {
	_, err := f.svc.DeleteService(&servicediscovery.DeleteServiceInput{
		Id: f.ID,
	})

	return err
}

func (f *ServiceDiscoveryService) String() string {
	return *f.ID
}
