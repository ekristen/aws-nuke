package resources

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/s3files"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const S3FilesAccessPointResource = "S3FilesAccessPoint"

func init() {
	registry.Register(&registry.Registration{
		Name:     S3FilesAccessPointResource,
		Scope:    nuke.Account,
		Resource: &S3FilesAccessPoint{},
		Lister:   &S3FilesAccessPointLister{},
	})
}

type S3FilesAccessPointLister struct {
	svc S3FilesAPI
}

func (l *S3FilesAccessPointLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
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
		params := &s3files.ListAccessPointsInput{
			FileSystemId: fsID,
		}

		for {
			res, err := l.svc.ListAccessPoints(ctx, params)
			if err != nil {
				return nil, err
			}

			for _, p := range res.AccessPoints {
				resources = append(resources, &S3FilesAccessPoint{
					svc:          l.svc,
					ID:           p.AccessPointId,
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

type S3FilesAccessPoint struct {
	svc          S3FilesAPI
	ID           *string `description:"The ID of the S3 file system access point"`
	FileSystemID *string `description:"The ID of the S3 file system that this access point belongs to"`
}

func (r *S3FilesAccessPoint) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteAccessPoint(ctx, &s3files.DeleteAccessPointInput{
		AccessPointId: r.ID,
	})
	return err
}

func (r *S3FilesAccessPoint) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *S3FilesAccessPoint) String() string {
	return *r.ID
}
