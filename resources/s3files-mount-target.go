package resources

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/s3files"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const S3FilesMountTargetResource = "S3FilesMountTarget"

func init() {
	registry.Register(&registry.Registration{
		Name:     S3FilesMountTargetResource,
		Scope:    nuke.Account,
		Resource: &S3FilesMountTarget{},
		Lister:   &S3FilesMountTargetLister{},
	})
}

type S3FilesMountTargetLister struct {
	svc S3FilesAPI
}

func (l *S3FilesMountTargetLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	var resources []resource.Resource

	if l.svc == nil {
		l.svc = s3files.NewFromConfig(*opts.Config)
	}

	fsIDs, err := listS3FileSystems(ctx, l.svc)
	if err != nil {
		return nil, err
	}

	for _, fsID := range fsIDs {
		params := &s3files.ListMountTargetsInput{
			FileSystemId: fsID,
		}

		for {
			res, err := l.svc.ListMountTargets(ctx, params)
			if err != nil {
				return nil, err
			}

			for _, p := range res.MountTargets {
				resources = append(resources, &S3FilesMountTarget{
					svc:          l.svc,
					ID:           p.MountTargetId,
					FileSystemID: fsID,
				})
			}

			if res.NextToken == nil {
				break
			}

			params.NextToken = res.NextToken
		}
	}

	return resources, nil
}

type S3FilesMountTarget struct {
	svc          S3FilesAPI
	ID           *string `description:"The ID of the S3 file system mount target"`
	FileSystemID *string `description:"The ID of the S3 file system that this mount target belongs to"`
}

func (r *S3FilesMountTarget) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteMountTarget(ctx, &s3files.DeleteMountTargetInput{
		MountTargetId: r.ID,
	})
	return err
}

func (r *S3FilesMountTarget) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *S3FilesMountTarget) String() string {
	return *r.ID
}
