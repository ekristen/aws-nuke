package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/simpledb"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const SimpleDBDomainResource = "SimpleDBDomain"

func init() {
	registry.Register(&registry.Registration{
		Name:   SimpleDBDomainResource,
		Scope:  nuke.Account,
		Lister: &SimpleDBDomainLister{},
	})
}

type SimpleDBDomainLister struct{}

func (l *SimpleDBDomainLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := simpledb.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &simpledb.ListDomainsInput{
		MaxNumberOfDomains: aws.Int64(100),
	}

	for {
		output, err := svc.ListDomains(params)
		if err != nil {
			return nil, err
		}

		for _, domainName := range output.DomainNames {
			resources = append(resources, &SimpleDBDomain{
				svc:        svc,
				domainName: domainName,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type SimpleDBDomain struct {
	svc        *simpledb.SimpleDB
	domainName *string
}

func (f *SimpleDBDomain) Remove(_ context.Context) error {
	_, err := f.svc.DeleteDomain(&simpledb.DeleteDomainInput{
		DomainName: f.domainName,
	})

	return err
}

func (f *SimpleDBDomain) String() string {
	return *f.domainName
}
