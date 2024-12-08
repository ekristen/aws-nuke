package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/codeartifact"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const CodeArtifactDomainResource = "CodeArtifactDomain"

func init() {
	registry.Register(&registry.Registration{
		Name:     CodeArtifactDomainResource,
		Scope:    nuke.Account,
		Resource: &CodeArtifactDomain{},
		Lister:   &CodeArtifactDomainLister{},
	})
}

type CodeArtifactDomainLister struct{}

func (l *CodeArtifactDomainLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := codeartifact.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &codeartifact.ListDomainsInput{}

	for {
		resp, err := svc.ListDomains(params)
		if err != nil {
			return nil, err
		}

		for _, domain := range resp.Domains {
			desc, err := svc.DescribeDomain(&codeartifact.DescribeDomainInput{Domain: domain.Name})
			if err != nil {
				return nil, err
			}

			resources = append(resources, &CodeArtifactDomain{
				svc:  svc,
				name: domain.Name,
				tags: GetDomainTags(svc, desc.Domain.Arn),
			})
		}

		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return resources, nil
}

func GetDomainTags(svc *codeartifact.CodeArtifact, arn *string) map[string]*string {
	tags := map[string]*string{}

	resp, _ := svc.ListTagsForResource(&codeartifact.ListTagsForResourceInput{ResourceArn: arn})
	for _, tag := range resp.Tags {
		tags[*tag.Key] = tag.Value
	}

	return tags
}

type CodeArtifactDomain struct {
	svc  *codeartifact.CodeArtifact
	name *string
	tags map[string]*string
}

func (d *CodeArtifactDomain) Remove(_ context.Context) error {
	_, err := d.svc.DeleteDomain(&codeartifact.DeleteDomainInput{
		Domain: d.name,
	})
	return err
}

func (d *CodeArtifactDomain) String() string {
	return *d.name
}

func (d *CodeArtifactDomain) Properties() types.Properties {
	properties := types.NewProperties()
	for key, tag := range d.tags {
		properties.SetTag(&key, tag)
	}
	properties.Set("Name", d.name)
	return properties
}
