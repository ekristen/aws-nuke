package resources

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/service/comprehend"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const ComprehendEntityRecognizerResource = "ComprehendEntityRecognizer"

func init() {
	registry.Register(&registry.Registration{
		Name:   ComprehendEntityRecognizerResource,
		Scope:  nuke.Account,
		Lister: &ComprehendEntityRecognizerLister{},
	})
}

type ComprehendEntityRecognizerLister struct{}

func (l *ComprehendEntityRecognizerLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := comprehend.New(opts.Session)

	params := &comprehend.ListEntityRecognizersInput{}
	resources := make([]resource.Resource, 0)

	for {
		resp, err := svc.ListEntityRecognizers(params)
		if err != nil {
			return nil, err
		}
		for _, entityRecognizer := range resp.EntityRecognizerPropertiesList {
			resources = append(resources, &ComprehendEntityRecognizer{
				svc:              svc,
				entityRecognizer: entityRecognizer,
			})
		}

		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return resources, nil
}

type ComprehendEntityRecognizer struct {
	svc              *comprehend.Comprehend
	entityRecognizer *comprehend.EntityRecognizerProperties
}

func (ce *ComprehendEntityRecognizer) Remove(_ context.Context) error {
	switch *ce.entityRecognizer.Status {
	case "IN_ERROR":
		fallthrough
	case "TRAINED":
		{
			logrus.Infof("ComprehendEntityRecognizer deleteEntityRecognizer arn=%s status=%s", *ce.entityRecognizer.EntityRecognizerArn, *ce.entityRecognizer.Status)
			_, err := ce.svc.DeleteEntityRecognizer(&comprehend.DeleteEntityRecognizerInput{
				EntityRecognizerArn: ce.entityRecognizer.EntityRecognizerArn,
			})
			return err
		}
	case "SUBMITTED":
		fallthrough
	case "TRAINING":
		{
			logrus.Infof("ComprehendEntityRecognizer stopTrainingEntityRecognizer arn=%s status=%s", *ce.entityRecognizer.EntityRecognizerArn, *ce.entityRecognizer.Status)
			_, err := ce.svc.StopTrainingEntityRecognizer(&comprehend.StopTrainingEntityRecognizerInput{
				EntityRecognizerArn: ce.entityRecognizer.EntityRecognizerArn,
			})
			return err
		}
	default:
		{
			logrus.Infof("ComprehendEntityRecognizer already deleting arn=%s status=%s", *ce.entityRecognizer.EntityRecognizerArn, *ce.entityRecognizer.Status)
			return nil
		}
	}
}

func (ce *ComprehendEntityRecognizer) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("LanguageCode", ce.entityRecognizer.LanguageCode)
	properties.Set("EntityRecognizerArn", ce.entityRecognizer.EntityRecognizerArn)

	return properties
}

func (ce *ComprehendEntityRecognizer) String() string {
	return *ce.entityRecognizer.EntityRecognizerArn
}
