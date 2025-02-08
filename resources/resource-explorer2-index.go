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

func (l *ResourceExplorer2IndexLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := resourceexplorer2.New(opts.Session)
	var resources []resource.Resource

	params := &resourceexplorer2.ListIndexesInput{
		Regions:    aws.StringSlice([]string{opts.Region.Name}),
		MaxResults: aws.Int64(100),
	}

	for {
		output, err := svc.ListIndexes(params)
		if err != nil {
			return nil, err
		}

		for _, index := range output.Indexes {
			var tags map[string]*string
			tagResp, err := svc.ListTagsForResource(
				&resourceexplorer2.ListTagsForResourceInput{
					ResourceArn: index.Arn,
				})
			if err != nil {
				logrus.WithError(err).Error("unable to list tags for resource")
			}
			if tagResp != nil {
				tags = tagResp.Tags
			}

			resources = append(resources, &ResourceExplorer2Index{
				svc:  svc,
				ARN:  index.Arn,
				Type: index.Type,
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

type ResourceExplorer2Index struct {
	svc  *resourceexplorer2.ResourceExplorer2
	ARN  *string
	Type *string
	Tags map[string]*string
}

func (r *ResourceExplorer2Index) Remove(_ context.Context) error {
	_, err := r.svc.DeleteIndex(&resourceexplorer2.DeleteIndexInput{
		Arn: r.ARN,
	})

	return err
}

func (r *ResourceExplorer2Index) String() string {
	return *r.ARN
}

func (r *ResourceExplorer2Index) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}
