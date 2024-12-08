package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/resourceexplorer2"
	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
)

const ResourceExplorer2IndexResource = "ResourceExplorer2Index"

func init() {
	registry.Register(&registry.Registration{
		Name:     ResourceExplorer2IndexResource,
		Scope:    nuke.Account,
		Resource: &ResourceExplorer2Index{},
		Lister:   &ResourceExplorer2IndexLister{},
	})
}

type ResourceExplorer2IndexLister struct{}

type ResourceExplorer2Index struct {
	svc      *resourceexplorer2.ResourceExplorer2
	indexArn *string
}

func (l *ResourceExplorer2IndexLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := resourceexplorer2.New(opts.Session)
	var resources []resource.Resource

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

func (f *ResourceExplorer2Index) Remove(_ context.Context) error {
	_, err := f.svc.DeleteIndex(&resourceexplorer2.DeleteIndexInput{
		Arn: f.indexArn,
	})

	return err
}

func (f *ResourceExplorer2Index) String() string {
	return *f.indexArn
}
