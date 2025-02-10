package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/resourceexplorer2"
	"github.com/sirupsen/logrus"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const ResourceExplorer2ViewResource = "ResourceExplorer2View"

func init() {
	registry.Register(&registry.Registration{
		Name:     ResourceExplorer2ViewResource,
		Scope:    nuke.Account,
		Resource: &ResourceExplorer2View{},
		Lister:   &ResourceExplorer2ViewLister{},
	})
}

type ResourceExplorer2ViewLister struct{}

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
			var tags map[string]*string
			tagResp, err := svc.ListTagsForResource(
				&resourceexplorer2.ListTagsForResourceInput{
					ResourceArn: view,
				})
			if err != nil {
				logrus.WithError(err).Error("unable to list tags for resource")
			}
			if tagResp != nil {
				tags = tagResp.Tags
			}

			resources = append(resources, &ResourceExplorer2View{
				svc:  svc,
				ARN:  view,
				Tags: tags,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.SetNextToken(aws.StringValue(output.NextToken))
	}

	return resources, nil
}

type ResourceExplorer2View struct {
	svc  *resourceexplorer2.ResourceExplorer2
	ARN  *string `description:"The ARN of the Resource Explorer View"`
	Tags map[string]*string
}

func (r *ResourceExplorer2View) Remove(_ context.Context) error {
	_, err := r.svc.DeleteView(&resourceexplorer2.DeleteViewInput{
		ViewArn: r.ARN,
	})

	return err
}

func (r *ResourceExplorer2View) String() string {
	return *r.ARN
}

func (r *ResourceExplorer2View) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}
