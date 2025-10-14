package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"         //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/fsx" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const FSxFileSystemResource = "FSxFileSystem"

func init() {
	registry.Register(&registry.Registration{
		Name:     FSxFileSystemResource,
		Scope:    nuke.Account,
		Resource: &FSxFileSystem{},
		Lister:   &FSxFileSystemLister{},
	})
}

type FSxFileSystemLister struct{}

func (l *FSxFileSystemLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := fsx.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &fsx.DescribeFileSystemsInput{
		MaxResults: aws.Int64(100),
	}

	for {
		resp, err := svc.DescribeFileSystems(params)
		if err != nil {
			return nil, err
		}

		for _, filesystem := range resp.FileSystems {
			resources = append(resources, &FSxFileSystem{
				svc:        svc,
				filesystem: filesystem,
			})
		}
		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}
	return resources, nil
}

type FSxFileSystem struct {
	svc        *fsx.FSx
	filesystem *fsx.FileSystem
}

func (f *FSxFileSystem) Remove(_ context.Context) error {
	_, err := f.svc.DeleteFileSystem(&fsx.DeleteFileSystemInput{
		FileSystemId: f.filesystem.FileSystemId,
	})

	return err
}

func (f *FSxFileSystem) Properties() types.Properties {
	properties := types.NewProperties()
	for _, tagValue := range f.filesystem.Tags {
		properties.SetTag(tagValue.Key, tagValue.Value)
	}
	properties.Set("Type", f.filesystem.FileSystemType)
	return properties
}

func (f *FSxFileSystem) String() string {
	return *f.filesystem.FileSystemId
}
