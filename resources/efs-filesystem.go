package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/efs"

	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const EFSFileSystemResource = "EFSFileSystem"

func init() {
	resource.Register(&resource.Registration{
		Name:   EFSFileSystemResource,
		Scope:  nuke.Account,
		Lister: &EFSFileSystemLister{},
	})
}

type EFSFileSystemLister struct{}

func (l *EFSFileSystemLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := efs.New(opts.Session)

	resp, err := svc.DescribeFileSystems(nil)
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, 0)
	for _, fs := range resp.FileSystems {
		lto, err := svc.ListTagsForResource(&efs.ListTagsForResourceInput{ResourceId: fs.FileSystemId})
		if err != nil {
			return nil, err
		}
		resources = append(resources, &EFSFileSystem{
			svc:     svc,
			id:      *fs.FileSystemId,
			name:    *fs.CreationToken,
			tagList: lto.Tags,
		})

	}

	return resources, nil
}

type EFSFileSystem struct {
	svc     *efs.EFS
	id      string
	name    string
	tagList []*efs.Tag
}

func (e *EFSFileSystem) Remove(_ context.Context) error {
	_, err := e.svc.DeleteFileSystem(&efs.DeleteFileSystemInput{
		FileSystemId: &e.id,
	})

	return err
}

func (e *EFSFileSystem) Properties() types.Properties {
	properties := types.NewProperties()
	for _, t := range e.tagList {
		properties.SetTag(t.Key, t.Value)
	}
	return properties
}

func (e *EFSFileSystem) String() string {
	return e.name
}
