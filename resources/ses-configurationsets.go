package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ses"

	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const SESConfigurationSetResource = "SESConfigurationSet"

func init() {
	resource.Register(resource.Registration{
		Name:   SESConfigurationSetResource,
		Scope:  nuke.Account,
		Lister: &SESConfigurationSetLister{},
	})
}

type SESConfigurationSetLister struct{}

func (l *SESConfigurationSetLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := ses.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &ses.ListConfigurationSetsInput{
		MaxItems: aws.Int64(100),
	}

	for {
		output, err := svc.ListConfigurationSets(params)
		if err != nil {
			return nil, err
		}

		for _, configurationSet := range output.ConfigurationSets {
			resources = append(resources, &SESConfigurationSet{
				svc:  svc,
				name: configurationSet.Name,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type SESConfigurationSet struct {
	svc  *ses.SES
	name *string
}

func (f *SESConfigurationSet) Remove(_ context.Context) error {
	_, err := f.svc.DeleteConfigurationSet(&ses.DeleteConfigurationSetInput{
		ConfigurationSetName: f.name,
	})

	return err
}

func (f *SESConfigurationSet) String() string {
	return *f.name
}
