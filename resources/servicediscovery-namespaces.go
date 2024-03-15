package resources

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicediscovery"
	"github.com/aws/aws-sdk-go/service/servicediscovery/servicediscoveryiface"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const ServiceDiscoveryNamespaceResource = "ServiceDiscoveryNamespace"

func init() {
	registry.Register(&registry.Registration{
		Name:   ServiceDiscoveryNamespaceResource,
		Scope:  nuke.Account,
		Lister: &ServiceDiscoveryNamespaceLister{},
	})
}

type ServiceDiscoveryNamespaceLister struct {
	mockSvc servicediscoveryiface.ServiceDiscoveryAPI
}

func (l *ServiceDiscoveryNamespaceLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	var svc servicediscoveryiface.ServiceDiscoveryAPI
	if l.mockSvc != nil {
		svc = l.mockSvc
	} else {
		svc = servicediscovery.New(opts.Session)
	}

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
			var tags []*servicediscovery.Tag
			tagsOutput, err := svc.ListTagsForResource(&servicediscovery.ListTagsForResourceInput{
				ResourceARN: namespace.Arn,
			})
			if err != nil {
				logrus.WithError(err).Error("unable to list tags for namespace")
			}
			if tagsOutput.Tags != nil {
				tags = tagsOutput.Tags
			}

			resources = append(resources, &ServiceDiscoveryNamespace{
				svc:  svc,
				ID:   namespace.Id,
				tags: tags,
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
	svc  servicediscoveryiface.ServiceDiscoveryAPI
	ID   *string
	tags []*servicediscovery.Tag
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

func (f *ServiceDiscoveryNamespace) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("ID", f.ID)

	for _, tag := range f.tags {
		properties.SetTag(tag.Key, tag.Value)
	}

	return properties
}
