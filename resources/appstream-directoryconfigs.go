package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appstream"

	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

type AppStreamDirectoryConfig struct {
	svc  *appstream.AppStream
	name *string
}

const AppStreamDirectoryConfigResource = "AppStreamDirectoryConfig"

func init() {
	resource.Register(resource.Registration{
		Name:   AppStreamDirectoryConfigResource,
		Scope:  nuke.Account,
		Lister: &AppStreamDirectoryConfigLister{},
	})
}

type AppStreamDirectoryConfigLister struct{}

func (l *AppStreamDirectoryConfigLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := appstream.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &appstream.DescribeDirectoryConfigsInput{
		MaxResults: aws.Int64(100),
	}

	for {
		output, err := svc.DescribeDirectoryConfigs(params)
		if err != nil {
			return nil, err
		}

		for _, directoryConfig := range output.DirectoryConfigs {
			resources = append(resources, &AppStreamDirectoryConfig{
				svc:  svc,
				name: directoryConfig.DirectoryName,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

func (f *AppStreamDirectoryConfig) Remove(_ context.Context) error {
	_, err := f.svc.DeleteDirectoryConfig(&appstream.DeleteDirectoryConfigInput{
		DirectoryName: f.name,
	})

	return err
}

func (f *AppStreamDirectoryConfig) String() string {
	return *f.name
}
