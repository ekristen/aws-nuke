package resources

import (
	"context"

	"go.uber.org/ratelimit"

	"github.com/aws/aws-sdk-go/service/efs"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const EFSFileSystemResource = "EFSFileSystem"

func init() {
	registry.Register(&registry.Registration{
		Name:   EFSFileSystemResource,
		Scope:  nuke.Account,
		Lister: &EFSFileSystemLister{},
	})
}

type EFSFileSystemLister struct{}

func (l *EFSFileSystemLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := efs.New(opts.Session)
	resources := make([]resource.Resource, 0)

	// Note: AWS does not publish what the RPS is for the DescribeFileSystems API call
	// after a bit of trial and error it seems to be around 10 RPS
	describeRL := ratelimit.New(10)

	params := &efs.DescribeFileSystemsInput{}
	for {
		describeRL.Take()

		resp, err := svc.DescribeFileSystems(params)
		if err != nil {
			return nil, err
		}

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

		if resp.NextMarker == nil {
			break
		}

		params.Marker = resp.NextMarker
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
