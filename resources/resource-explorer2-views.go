package resources

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/resourceexplorer2"
)

type ResourceExplorer2View struct {
	svc     *resourceexplorer2.ResourceExplorer2
	viewArn *string
}

func init() {
	register("ResourceExplorer2View", ResourceExplorer2Views)
}

func ResourceExplorer2Views(sess *session.Session) ([]Resource, error) {
	svc := resourceexplorer2.New(sess)
	var resources []Resource

	params := &resourceexplorer2.ListViewsInput{}

	for {
		output, err := svc.ListViews(params)
		if err != nil {
			return nil, err
		}

		for _, view := range output.Views {
			resources = append(resources, &ResourceExplorer2View{
				svc:     svc,
				viewArn: view,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.SetNextToken(aws.StringValue(output.NextToken))
	}

	return resources, nil
}

func (f *ResourceExplorer2View) Remove() error {
	_, err := f.svc.DeleteView(&resourceexplorer2.DeleteViewInput{
		ViewArn: f.viewArn,
	})

	return err
}

func (f *ResourceExplorer2View) String() string {
	return *f.viewArn
}
