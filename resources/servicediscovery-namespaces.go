package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicediscovery"

	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const ServiceDiscoveryNamespaceResource = "ServiceDiscoveryNamespace"

func init() {
	resource.Register(&resource.Registration{
		Name:   ServiceDiscoveryNamespaceResource,
		Scope:  nuke.Account,
		Lister: &ServiceDiscoveryNamespaceLister{},
	})
}

type ServiceDiscoveryNamespaceLister struct{}

func (l *ServiceDiscoveryNamespaceLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := servicediscovery.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &servicediscovery.ListNamespacesInput{
		MaxResults: aws.Int64(100),
	}

	// Collect all services, using separate for loop
	// due to multi-service pagination issues
	for {
		output, err := svc.ListNamespaces(params)
		if err != nil {
			return nil, err
		}

		for _, namespace := range output.Namespaces {
			resources = append(resources, &ServiceDiscoveryNamespace{
				svc: svc,
				ID:  namespace.Id,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type ServiceDiscoveryNamespace struct {
	svc *servicediscovery.ServiceDiscovery
	ID  *string
}

func (f *ServiceDiscoveryNamespace) Remove(_ context.Context) error {
	_, err := f.svc.DeleteNamespace(&servicediscovery.DeleteNamespaceInput{
		Id: f.ID,
	})

	return err
}

func (f *ServiceDiscoveryNamespace) String() string {
	return *f.ID
}
