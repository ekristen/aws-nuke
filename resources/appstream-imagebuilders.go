package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appstream"

	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

type AppStreamImageBuilder struct {
	svc  *appstream.AppStream
	name *string
}

const AppStreamImageBuilderResource = "AppStreamImageBuilder"

func init() {
	resource.Register(&resource.Registration{
		Name:   AppStreamImageBuilderResource,
		Scope:  nuke.Account,
		Lister: &AppStreamImageBuilderLister{},
	})
}

type AppStreamImageBuilderLister struct{}

func (l *AppStreamImageBuilderLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := appstream.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &appstream.DescribeImageBuildersInput{
		MaxResults: aws.Int64(100),
	}

	for {
		output, err := svc.DescribeImageBuilders(params)
		if err != nil {
			return nil, err
		}

		for _, imageBuilder := range output.ImageBuilders {
			resources = append(resources, &AppStreamImageBuilder{
				svc:  svc,
				name: imageBuilder.Name,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

func (f *AppStreamImageBuilder) Remove(_ context.Context) error {
	_, err := f.svc.DeleteImageBuilder(&appstream.DeleteImageBuilderInput{
		Name: f.name,
	})

	return err
}

func (f *AppStreamImageBuilder) String() string {
	return *f.name
}
