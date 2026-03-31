package resources

import (
	"context"
	stderrors "errors"

	"github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/comprehend"
	comptypes "github.com/aws/aws-sdk-go-v2/service/comprehend/types"

	"github.com/ekristen/libnuke/pkg/errors"
	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const ComprehendDocumentClassifierResource = "ComprehendDocumentClassifier"

func init() {
	registry.Register(&registry.Registration{
		Name:     ComprehendDocumentClassifierResource,
		Scope:    nuke.Account,
		Resource: &ComprehendDocumentClassifier{},
		Lister:   &ComprehendDocumentClassifierLister{},
	})
}

type ComprehendDocumentClassifierLister struct{}

func (l *ComprehendDocumentClassifierLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := comprehend.NewFromConfig(*opts.Config)
	resources := make([]resource.Resource, 0)

	paginator := comprehend.NewListDocumentClassifiersPaginator(svc, &comprehend.ListDocumentClassifiersInput{
		MaxResults: aws.Int32(100),
	})

	for paginator.HasMorePages() {
		resp, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for i := range resp.DocumentClassifierPropertiesList {
			resources = append(resources, &ComprehendDocumentClassifier{
				svc:                svc,
				documentClassifier: resp.DocumentClassifierPropertiesList[i],
			})
		}
	}

	return resources, nil
}

type ComprehendDocumentClassifier struct {
	svc                *comprehend.Client
	documentClassifier comptypes.DocumentClassifierProperties
}

func (ce *ComprehendDocumentClassifier) Remove(ctx context.Context) error {
	switch ce.documentClassifier.Status {
	case comptypes.ModelStatusInError, comptypes.ModelStatusTrained, comptypes.ModelStatusStopped:
		logrus.Infof("ComprehendDocumentClassifier deleteDocumentClassifier arn=%s status=%s",
			aws.ToString(ce.documentClassifier.DocumentClassifierArn), ce.documentClassifier.Status)

		_, err := ce.svc.DeleteDocumentClassifier(ctx, &comprehend.DeleteDocumentClassifierInput{
			DocumentClassifierArn: aws.String(aws.ToString(ce.documentClassifier.DocumentClassifierArn)),
		})
		return err
	case comptypes.ModelStatusSubmitted, comptypes.ModelStatusTraining:
		logrus.Infof("ComprehendDocumentClassifier stopTrainingDocumentClassifier arn=%s status=%s",
			aws.ToString(ce.documentClassifier.DocumentClassifierArn), ce.documentClassifier.Status)

		_, err := ce.svc.StopTrainingDocumentClassifier(ctx, &comprehend.StopTrainingDocumentClassifierInput{
			DocumentClassifierArn: aws.String(aws.ToString(ce.documentClassifier.DocumentClassifierArn)),
		})
		return err
	default:
		logrus.Infof("ComprehendDocumentClassifier already deleting arn=%s status=%s",
			aws.ToString(ce.documentClassifier.DocumentClassifierArn), ce.documentClassifier.Status)
		return nil
	}
}

func (ce *ComprehendDocumentClassifier) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("LanguageCode", string(ce.documentClassifier.LanguageCode))
	properties.Set("DocumentClassifierArn", aws.ToString(ce.documentClassifier.DocumentClassifierArn))

	return properties
}

func (ce *ComprehendDocumentClassifier) String() string {
	return aws.ToString(ce.documentClassifier.DocumentClassifierArn)
}

func (ce *ComprehendDocumentClassifier) HandleWait(ctx context.Context) error {
	resp, err := ce.svc.DescribeDocumentClassifier(ctx, &comprehend.DescribeDocumentClassifierInput{
		DocumentClassifierArn: aws.String(aws.ToString(ce.documentClassifier.DocumentClassifierArn)),
	})
	if err != nil {
		// Check if classifier no longer exists
		var rnf *comptypes.ResourceNotFoundException
		if stderrors.As(err, &rnf) {
			logrus.Info("ComprehendDocumentClassifier removed")
			return nil
		}
		return err
	}

	logrus.Infof("ComprehendDocumentClassifier arn=%s, has status=%s",
		aws.ToString(ce.documentClassifier.DocumentClassifierArn), resp.DocumentClassifierProperties.Status)

	switch resp.DocumentClassifierProperties.Status {
	case comptypes.ModelStatusDeleting:
		return errors.ErrWaitResource("document classifier is still deleting")
	case comptypes.ModelStatusStopped:
		logrus.Info("ComprehendDocumentClassifier stopped, attempting deletion")
		_, err := ce.svc.DeleteDocumentClassifier(ctx, &comprehend.DeleteDocumentClassifierInput{
			DocumentClassifierArn: aws.String(aws.ToString(ce.documentClassifier.DocumentClassifierArn)),
		})
		return err

	default:
		return errors.ErrWaitResource("document classifier not deleted yet")
	}
}
