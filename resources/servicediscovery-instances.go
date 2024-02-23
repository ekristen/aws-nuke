package resources

import (
	"context"

	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicediscovery"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const ServiceDiscoveryInstanceResource = "ServiceDiscoveryInstance"

func init() {
	registry.Register(&registry.Registration{
		Name:   ServiceDiscoveryInstanceResource,
		Scope:  nuke.Account,
		Lister: &ServiceDiscoveryInstanceLister{},
	})
}

type ServiceDiscoveryInstanceLister struct{}

func (l *ServiceDiscoveryInstanceLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := servicediscovery.New(opts.Session)
	resources := make([]resource.Resource, 0)
	var services []*servicediscovery.ServiceSummary

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

		services = append(services, output.Services...)

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	// collect instances for de-registration
	for _, service := range services {
		instanceParams := &servicediscovery.ListInstancesInput{
			ServiceId:  service.Id,
			MaxResults: aws.Int64(100),
		}

		output, err := svc.ListInstances(instanceParams)
		if err != nil {
			return nil, err
		}

		for _, instance := range output.Instances {
			resources = append(resources, &ServiceDiscoveryInstance{
				svc:        svc,
				serviceID:  service.Id,
				instanceID: instance.Id,
			})
		}
		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type ServiceDiscoveryInstance struct {
	svc        *servicediscovery.ServiceDiscovery
	serviceID  *string
	instanceID *string
}

func (f *ServiceDiscoveryInstance) Remove(_ context.Context) error {
	_, err := f.svc.DeregisterInstance(&servicediscovery.DeregisterInstanceInput{
		InstanceId: f.instanceID,
		ServiceId:  f.serviceID,
	})

	return err
}

func (f *ServiceDiscoveryInstance) String() string {
	return fmt.Sprintf("%s -> %s", *f.instanceID, *f.serviceID)
}
