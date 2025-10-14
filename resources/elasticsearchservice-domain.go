package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/elasticsearchservice" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const ESDomainResource = "ESDomain"

func init() {
	registry.Register(&registry.Registration{
		Name:     ESDomainResource,
		Scope:    nuke.Account,
		Resource: &ESDomain{},
		Lister:   &ESDomainLister{},
	})
}

type ESDomainLister struct{}

func (l *ESDomainLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := elasticsearchservice.New(opts.Session)

	params := &elasticsearchservice.ListDomainNamesInput{}
	resp, err := svc.ListDomainNames(params)
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, 0)
	for _, domain := range resp.DomainNames {
		dedo, err := svc.DescribeElasticsearchDomain(
			&elasticsearchservice.DescribeElasticsearchDomainInput{DomainName: domain.DomainName})
		if err != nil {
			return nil, err
		}
		lto, err := svc.ListTags(&elasticsearchservice.ListTagsInput{ARN: dedo.DomainStatus.ARN})
		if err != nil {
			return nil, err
		}
		resources = append(resources, &ESDomain{
			svc:        svc,
			domainName: domain.DomainName,
			tagList:    lto.TagList,
		})
	}

	return resources, nil
}

type ESDomain struct {
	svc        *elasticsearchservice.ElasticsearchService
	domainName *string
	tagList    []*elasticsearchservice.Tag
}

func (f *ESDomain) Remove(_ context.Context) error {
	_, err := f.svc.DeleteElasticsearchDomain(&elasticsearchservice.DeleteElasticsearchDomainInput{
		DomainName: f.domainName,
	})

	return err
}

func (f *ESDomain) Properties() types.Properties {
	properties := types.NewProperties()
	for _, t := range f.tagList {
		properties.SetTag(t.Key, t.Value)
	}
	return properties
}

func (f *ESDomain) String() string {
	return *f.domainName
}
