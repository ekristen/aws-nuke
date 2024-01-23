package resources

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/service/comprehend"

	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const ComprehendDocumentClassifierResource = "ComprehendDocumentClassifier"

func init() {
	resource.Register(&resource.Registration{
		Name:   ComprehendDocumentClassifierResource,
		Scope:  nuke.Account,
		Lister: &ComprehendDocumentClassifierLister{},
	})
}

type ComprehendDocumentClassifierLister struct{}

func (l *ComprehendDocumentClassifierLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := comprehend.New(opts.Session)

	params := &comprehend.ListDocumentClassifiersInput{}
	resources := make([]resource.Resource, 0)

	for {
		resp, err := svc.ListDocumentClassifiers(params)
		if err != nil {
			return nil, err
		}
		for _, documentClassifier := range resp.DocumentClassifierPropertiesList {
			resources = append(resources, &ComprehendDocumentClassifier{
				svc:                svc,
				documentClassifier: documentClassifier,
			})
		}

		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return resources, nil
}

type ComprehendDocumentClassifier struct {
	svc                *comprehend.Comprehend
	documentClassifier *comprehend.DocumentClassifierProperties
}

func (ce *ComprehendDocumentClassifier) Remove(_ context.Context) error {
	switch *ce.documentClassifier.Status {
	case "IN_ERROR":
		fallthrough
	case "TRAINED":
		{
			logrus.Infof("ComprehendDocumentClassifier deleteDocumentClassifier arn=%s status=%s", *ce.documentClassifier.DocumentClassifierArn, *ce.documentClassifier.Status)
			_, err := ce.svc.DeleteDocumentClassifier(&comprehend.DeleteDocumentClassifierInput{
				DocumentClassifierArn: ce.documentClassifier.DocumentClassifierArn,
			})
			return err
		}
	case "SUBMITTED":
		fallthrough
	case "TRAINING":
		{
			logrus.Infof("ComprehendDocumentClassifier stopTrainingDocumentClassifier arn=%s status=%s", *ce.documentClassifier.DocumentClassifierArn, *ce.documentClassifier.Status)
			_, err := ce.svc.StopTrainingDocumentClassifier(&comprehend.StopTrainingDocumentClassifierInput{
				DocumentClassifierArn: ce.documentClassifier.DocumentClassifierArn,
			})
			return err
		}
	default:
		{
			logrus.Infof("ComprehendDocumentClassifier already deleting arn=%s status=%s", *ce.documentClassifier.DocumentClassifierArn, *ce.documentClassifier.Status)
			return nil
		}
	}
}

func (ce *ComprehendDocumentClassifier) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("LanguageCode", ce.documentClassifier.LanguageCode)
	properties.Set("DocumentClassifierArn", ce.documentClassifier.DocumentClassifierArn)

	return properties
}

func (ce *ComprehendDocumentClassifier) String() string {
	return *ce.documentClassifier.DocumentClassifierArn
}
