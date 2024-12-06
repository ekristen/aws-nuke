package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ses"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const SESTemplateResource = "SESTemplate"

func init() {
	registry.Register(&registry.Registration{
		Name:     SESTemplateResource,
		Scope:    nuke.Account,
		Resource: &SESTemplate{},
		Lister:   &SESTemplateLister{},
	})
}

type SESTemplateLister struct{}

func (l *SESTemplateLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := ses.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &ses.ListTemplatesInput{
		MaxItems: aws.Int64(100),
	}

	for {
		output, err := svc.ListTemplates(params)
		if err != nil {
			return nil, err
		}

		for _, templateMetadata := range output.TemplatesMetadata {
			resources = append(resources, &SESTemplate{
				svc:  svc,
				name: templateMetadata.Name,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type SESTemplate struct {
	svc  *ses.SES
	name *string
}

func (f *SESTemplate) Remove(_ context.Context) error {
	_, err := f.svc.DeleteTemplate(&ses.DeleteTemplateInput{
		TemplateName: f.name,
	})

	return err
}

func (f *SESTemplate) String() string {
	return *f.name
}
