package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/cloudsearch"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const CloudSearchDomainResource = "CloudSearchDomain"

func init() {
	registry.Register(&registry.Registration{
		Name:   CloudSearchDomainResource,
		Scope:  nuke.Account,
		Lister: &CloudSearchDomainLister{},
	})
}

type CloudSearchDomainLister struct{}

func (l *CloudSearchDomainLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := cloudsearch.New(opts.Session)

	params := &cloudsearch.DescribeDomainsInput{}

	resp, err := svc.DescribeDomains(params)
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, 0)
	for _, domain := range resp.DomainStatusList {
		resources = append(resources, &CloudSearchDomain{
			svc:        svc,
			domainName: domain.DomainName,
		})
	}
	return resources, nil
}

type CloudSearchDomain struct {
	svc        *cloudsearch.CloudSearch
	domainName *string
}

func (f *CloudSearchDomain) Remove(_ context.Context) error {
	_, err := f.svc.DeleteDomain(&cloudsearch.DeleteDomainInput{
		DomainName: f.domainName,
	})

	return err
}

func (f *CloudSearchDomain) String() string {
	return *f.domainName
}
