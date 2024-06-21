package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const SSMParameterResource = "SSMParameter"

func init() {
	registry.Register(&registry.Registration{
		Name:   SSMParameterResource,
		Scope:  nuke.Account,
		Lister: &SSMParameterLister{},
	})
}

type SSMParameterLister struct{}

func (l *SSMParameterLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := ssm.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &ssm.DescribeParametersInput{
		MaxResults: aws.Int64(50),
	}

	for {
		output, err := svc.DescribeParameters(params)
		if err != nil {
			return nil, err
		}

		for _, parameter := range output.Parameters {
			tagParams := &ssm.ListTagsForResourceInput{
				ResourceId:   parameter.Name,
				ResourceType: aws.String(ssm.ResourceTypeForTaggingParameter),
			}

			tagResp, tagErr := svc.ListTagsForResource(tagParams)
			if tagErr != nil {
				return nil, tagErr
			}

			resources = append(resources, &SSMParameter{
				svc:  svc,
				name: parameter.Name,
				tags: tagResp.TagList,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type SSMParameter struct {
	svc  *ssm.SSM
	name *string
	tags []*ssm.Tag
}

func (f *SSMParameter) Remove(_ context.Context) error {
	_, err := f.svc.DeleteParameter(&ssm.DeleteParameterInput{
		Name: f.name,
	})

	return err
}

func (f *SSMParameter) String() string {
	return *f.name
}

func (f *SSMParameter) Properties() types.Properties {
	properties := types.NewProperties()
	for _, tag := range f.tags {
		properties.SetTag(tag.Key, tag.Value)
	}
	properties.
		Set("Name", f.name)
	return properties
}
