package resources

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/resourceexplorer2"
)

type ResourceExplorer2Index struct {
	svc      *resourceexplorer2.ResourceExplorer2
	indexArn *string
}

func init() {
	register("ResourceExplorer2Index", ResourceExplorer2Indexes)
}

func ResourceExplorer2Indexes(sess *session.Session) ([]Resource, error) {
	svc := resourceexplorer2.New(sess)
	var resources []Resource

	params := &resourceexplorer2.ListIndexesInput{}

	for {
		output, err := svc.ListIndexes(params)
		if err != nil {
			return nil, err
		}

		for _, index := range output.Indexes {
			resources = append(resources, &ResourceExplorer2Index{
				svc:      svc,
				indexArn: index.Arn,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.SetNextToken(aws.StringValue(output.NextToken))
	}

	return resources, nil
}

func (f *ResourceExplorer2Index) Remove() error {
	_, err := f.svc.DeleteIndex(&resourceexplorer2.DeleteIndexInput{
		Arn: f.indexArn,
	})

	return err
}

func (f *ResourceExplorer2Index) String() string {
	return *f.indexArn
}
