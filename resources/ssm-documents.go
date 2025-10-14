package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"         //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/ssm" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const SSMDocumentResource = "SSMDocument"

func init() {
	registry.Register(&registry.Registration{
		Name:     SSMDocumentResource,
		Scope:    nuke.Account,
		Resource: &SSMDocument{},
		Lister:   &SSMDocumentLister{},
	})
}

type SSMDocumentLister struct{}

func (l *SSMDocumentLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := ssm.New(opts.Session)
	resources := make([]resource.Resource, 0)

	documentKeyFilter := []*ssm.DocumentKeyValuesFilter{
		{
			Key:    aws.String("Owner"),
			Values: []*string{aws.String("Self")},
		},
	}

	params := &ssm.ListDocumentsInput{
		MaxResults: aws.Int64(50),
		Filters:    documentKeyFilter,
	}

	for {
		output, err := svc.ListDocuments(params)
		if err != nil {
			return nil, err
		}

		for _, documentIdentifier := range output.DocumentIdentifiers {
			resources = append(resources, &SSMDocument{
				svc:  svc,
				name: documentIdentifier.Name,
				tags: documentIdentifier.Tags,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type SSMDocument struct {
	svc  *ssm.SSM
	name *string
	tags []*ssm.Tag
}

func (f *SSMDocument) Remove(_ context.Context) error {
	_, err := f.svc.DeleteDocument(&ssm.DeleteDocumentInput{
		Name: f.name,
	})

	return err
}

func (f *SSMDocument) Properties() types.Properties {
	properties := types.NewProperties()
	for _, tagValue := range f.tags {
		properties.SetTag(tagValue.Key, tagValue.Value)
	}
	properties.Set("Name", f.name)
	return properties
}

func (f *SSMDocument) String() string {
	return *f.name
}
