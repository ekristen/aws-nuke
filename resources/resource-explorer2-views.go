package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/resourceexplorer2"
	"github.com/ekristen/aws-nuke/pkg/nuke"
	"github.com/ekristen/libnuke/pkg/resource"
)

const ResourceExplorer2ViewResource = "ResourceExplorer2View"

func init() {
	resource.Register(&resource.Registration{
		Name:   ResourceExplorer2ViewResource,
		Scope:  nuke.Account,
		Lister: &ResourceExplorer2ViewLister{},
	})
}

type ResourceExplorer2ViewLister struct{}

type ResourceExplorer2View struct {
	svc     *resourceexplorer2.ResourceExplorer2
	viewArn *string
}

func (l *ResourceExplorer2ViewLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := resourceexplorer2.New(opts.Session)
	var resources []resource.Resource

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

func (f *ResourceExplorer2View) Remove(_ context.Context) error {
	_, err := f.svc.DeleteView(&resourceexplorer2.DeleteViewInput{
		ViewArn: f.viewArn,
	})

	return err
}

func (f *ResourceExplorer2View) String() string {
	return *f.viewArn
}
