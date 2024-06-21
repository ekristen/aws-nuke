package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/glue"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const GlueClassifierResource = "GlueClassifier"

func init() {
	registry.Register(&registry.Registration{
		Name:   GlueClassifierResource,
		Scope:  nuke.Account,
		Lister: &GlueClassifierLister{},
	})
}

type GlueClassifierLister struct{}

func (l *GlueClassifierLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := glue.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &glue.GetClassifiersInput{
		MaxResults: aws.Int64(100),
	}

	for {
		output, err := svc.GetClassifiers(params)
		if err != nil {
			return nil, err
		}

		for _, classifier := range output.Classifiers {
			switch {
			case classifier.GrokClassifier != nil:
				resources = append(resources, &GlueClassifier{
					svc:  svc,
					name: classifier.GrokClassifier.Name,
				})
			case classifier.JsonClassifier != nil:
				resources = append(resources, &GlueClassifier{
					svc:  svc,
					name: classifier.JsonClassifier.Name,
				})
			case classifier.XMLClassifier != nil:
				resources = append(resources, &GlueClassifier{
					svc:  svc,
					name: classifier.XMLClassifier.Name,
				})
			}
		}
		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type GlueClassifier struct {
	svc  *glue.Glue
	name *string
}

func (f *GlueClassifier) Remove(_ context.Context) error {
	_, err := f.svc.DeleteClassifier(&glue.DeleteClassifierInput{
		Name: f.name,
	})

	return err
}

func (f *GlueClassifier) String() string {
	return *f.name
}
