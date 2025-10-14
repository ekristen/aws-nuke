package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"                 //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/rekognition" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const RekognitionProjectResource = "RekognitionProject"

func init() {
	registry.Register(&registry.Registration{
		Name:     RekognitionProjectResource,
		Scope:    nuke.Account,
		Resource: &RekognitionProject{},
		Lister:   &RekognitionProjectLister{},
	})
}

type RekognitionProjectLister struct{}

func (l *RekognitionProjectLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
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
			resources = append(resources, &RekognitionProject{
				svc: svc,
				ARN: project.ProjectArn,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type RekognitionProject struct {
	svc *rekognition.Rekognition
	ARN *string
}

func (r *RekognitionProject) Remove(_ context.Context) error {
	_, err := r.svc.DeleteProject(&rekognition.DeleteProjectInput{
		ProjectArn: r.ARN,
	})

	return err
}

func (r *RekognitionProject) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *RekognitionProject) String() string {
	return *r.ARN
}
