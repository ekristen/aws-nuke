package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/mediaconvert"

	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const MediaConvertJobTemplateResource = "MediaConvertJobTemplate"

func init() {
	resource.Register(resource.Registration{
		Name:   MediaConvertJobTemplateResource,
		Scope:  nuke.Account,
		Lister: &MediaConvertJobTemplateLister{},
	})
}

type MediaConvertJobTemplateLister struct{}

func (l *MediaConvertJobTemplateLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := mediaconvert.New(opts.Session)
	resources := make([]resource.Resource, 0)
	var mediaEndpoint *string

	output, err := svc.DescribeEndpoints(&mediaconvert.DescribeEndpointsInput{})
	if err != nil {
		return nil, err
	}

	for _, endpoint := range output.Endpoints {
		mediaEndpoint = endpoint.Url
	}

	// Update svc to use custom media endpoint
	svc.Endpoint = *mediaEndpoint

	params := &mediaconvert.ListJobTemplatesInput{
		MaxResults: aws.Int64(20),
	}

	for {
		output, err := svc.ListJobTemplates(params)
		if err != nil {
			return nil, err
		}

		for _, jobTemplate := range output.JobTemplates {
			resources = append(resources, &MediaConvertJobTemplate{
				svc:  svc,
				name: jobTemplate.Name,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type MediaConvertJobTemplate struct {
	svc  *mediaconvert.MediaConvert
	name *string
}

func (f *MediaConvertJobTemplate) Remove(_ context.Context) error {
	_, err := f.svc.DeleteJobTemplate(&mediaconvert.DeleteJobTemplateInput{
		Name: f.name,
	})

	return err
}

func (f *MediaConvertJobTemplate) String() string {
	return *f.name
}
