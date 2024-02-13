package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directoryservice"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const DirectoryServiceDirectoryResource = "DirectoryServiceDirectory"

func init() {
	registry.Register(&registry.Registration{
		Name:   DirectoryServiceDirectoryResource,
		Scope:  nuke.Account,
		Lister: &DirectoryServiceDirectoryLister{},
	})
}

type DirectoryServiceDirectoryLister struct{}

func (l *DirectoryServiceDirectoryLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := directoryservice.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &directoryservice.DescribeDirectoriesInput{
		Limit: aws.Int64(100),
	}

	for {
		resp, err := svc.DescribeDirectories(params)
		if err != nil {
			return nil, err
		}

		for _, directory := range resp.DirectoryDescriptions {
			resources = append(resources, &DirectoryServiceDirectory{
				svc:         svc,
				directoryID: directory.DirectoryId,
			})
		}

		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return resources, nil
}

type DirectoryServiceDirectory struct {
	svc         *directoryservice.DirectoryService
	directoryID *string
}

func (f *DirectoryServiceDirectory) Remove(_ context.Context) error {
	_, err := f.svc.DeleteDirectory(&directoryservice.DeleteDirectoryInput{
		DirectoryId: f.directoryID,
	})

	return err
}

func (f *DirectoryServiceDirectory) String() string {
	return *f.directoryID
}
