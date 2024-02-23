package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/clouddirectory"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const CloudDirectoryDirectoryResource = "CloudDirectoryDirectory"

func init() {
	registry.Register(&registry.Registration{
		Name:   CloudDirectoryDirectoryResource,
		Scope:  nuke.Account,
		Lister: &CloudDirectoryDirectoryLister{},
	})
}

type CloudDirectoryDirectoryLister struct{}

func (l *CloudDirectoryDirectoryLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := clouddirectory.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &clouddirectory.ListDirectoriesInput{
		MaxResults: aws.Int64(30),
		State:      aws.String("ENABLED"),
	}

	for {
		resp, err := svc.ListDirectories(params)
		if err != nil {
			return nil, err
		}

		for _, directory := range resp.Directories {
			resources = append(resources, &CloudDirectoryDirectory{
				svc:          svc,
				directoryARN: directory.DirectoryArn,
			})
		}

		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return resources, nil
}

type CloudDirectoryDirectory struct {
	svc          *clouddirectory.CloudDirectory
	directoryARN *string
}

func (f *CloudDirectoryDirectory) Remove(_ context.Context) error {
	_, err := f.svc.DisableDirectory(&clouddirectory.DisableDirectoryInput{
		DirectoryArn: f.directoryARN,
	})

	if err == nil {
		_, err = f.svc.DeleteDirectory(&clouddirectory.DeleteDirectoryInput{
			DirectoryArn: f.directoryARN,
		})
	}

	return err
}

func (f *CloudDirectoryDirectory) String() string {
	return *f.directoryARN
}
