package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"

	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const SSMDocumentResource = "SSMDocument"

func init() {
	resource.Register(&resource.Registration{
		Name:   SSMDocumentResource,
		Scope:  nuke.Account,
		Lister: &SSMDocumentLister{},
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
}

func (f *SSMDocument) Remove(_ context.Context) error {
	_, err := f.svc.DeleteDocument(&ssm.DeleteDocumentInput{
		Name: f.name,
	})

	return err
}

func (f *SSMDocument) String() string {
	return *f.name
}
