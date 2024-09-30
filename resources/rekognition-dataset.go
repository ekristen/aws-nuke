package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rekognition"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const RekognitionDatasetResource = "RekognitionDataset"

func init() {
	registry.Register(&registry.Registration{
		Name:   RekognitionDatasetResource,
		Scope:  nuke.Account,
		Lister: &RekognitionDatasetLister{},
	})
}

type RekognitionDatasetLister struct{}

func (l *RekognitionDatasetLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := rekognition.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &rekognition.DescribeProjectsInput{
		MaxResults: aws.Int64(100),
	}

	for {
		output, err := svc.DescribeProjects(params)
		if err != nil {
			return nil, err
		}

		for _, project := range output.ProjectDescriptions {
			for _, dataset := range project.Datasets {
				resources = append(resources, &RekognitionDataset{
					svc: svc,
					ARN: dataset.DatasetArn,
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

type RekognitionDataset struct {
	svc *rekognition.Rekognition
	ARN *string
}

func (r *RekognitionDataset) Remove(_ context.Context) error {
	_, err := r.svc.DeleteDataset(&rekognition.DeleteDatasetInput{
		DatasetArn: r.ARN,
	})

	return err
}

func (r *RekognitionDataset) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *RekognitionDataset) String() string {
	return *r.ARN
}
