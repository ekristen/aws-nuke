package resources

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/s3files"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const S3FilesFileSystemResource = "S3FilesFileSystem"

func init() {
	registry.Register(&registry.Registration{
		Name:     S3FilesFileSystemResource,
		Scope:    nuke.Account,
		Resource: &S3FilesFileSystem{},
		Lister:   &S3FilesFileSystemLister{},
		DependsOn: []string{
			S3FilesMountTargetResource,
			S3FilesAccessPointResource,
		},
	})
}

type S3FilesFileSystemLister struct {
	svc S3FilesAPI
}

func (l *S3FilesFileSystemLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	var resources []resource.Resource

	if l.svc == nil {
		l.svc = s3files.NewFromConfig(*opts.Config)
	}

	params := &s3files.ListFileSystemsInput{}

	for {
		res, err := l.svc.ListFileSystems(ctx, params)
		if err != nil {
			return nil, err
		}

		for _, p := range res.FileSystems {
			resources = append(resources, &S3FilesFileSystem{
				svc:  l.svc,
				ID:   p.FileSystemId,
				Name: p.Name,
			})
		}

		if res.NextToken == nil {
			break
		}

		params.NextToken = res.NextToken
	}

	return resources, nil
}

type S3FilesFileSystem struct {
	svc  S3FilesAPI
	ID   *string `description:"The ID of the S3 file system"`
	Name *string `description:"The name of the S3 file system"`
}

func (r *S3FilesFileSystem) Remove(ctx context.Context) error {
	b := true
	_, err := r.svc.DeleteFileSystem(ctx, &s3files.DeleteFileSystemInput{
		FileSystemId: r.ID,
		ForceDelete:  &b,
	})
	return err
}

func (r *S3FilesFileSystem) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *S3FilesFileSystem) String() string {
	if r.Name != nil {
		return fmt.Sprintf("%s (%s)", *r.ID, *r.Name)
	} else {
		return *r.ID
	}
}

func listS3FileSystems(ctx context.Context, svc S3FilesAPI) ([]*string, error) {
	var fsIDs []*string

	for {
		params := &s3files.ListFileSystemsInput{}

		res, err := svc.ListFileSystems(ctx, params)
		if err != nil {
			return nil, err
		}
		for _, p := range res.FileSystems {
			fsIDs = append(fsIDs, p.FileSystemId)
		}

		if res.NextToken == nil {
			break
		}

		params.NextToken = res.NextToken
	}

	return fsIDs, nil
}
