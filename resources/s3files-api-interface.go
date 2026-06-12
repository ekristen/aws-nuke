package resources

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/s3files"
)

// S3FilesAPI is the subset of the s3files client surface used by the
// S3Files* resources. Defining it as an interface lets the listers and
// resources be exercised with a gomock-generated fake.
type S3FilesAPI interface {
	ListFileSystems(ctx context.Context, params *s3files.ListFileSystemsInput,
		optFns ...func(*s3files.Options)) (*s3files.ListFileSystemsOutput, error)
	DeleteFileSystem(ctx context.Context, params *s3files.DeleteFileSystemInput,
		optFns ...func(*s3files.Options)) (*s3files.DeleteFileSystemOutput, error)
	ListAccessPoints(ctx context.Context, params *s3files.ListAccessPointsInput,
		optFns ...func(*s3files.Options)) (*s3files.ListAccessPointsOutput, error)
	DeleteAccessPoint(ctx context.Context, params *s3files.DeleteAccessPointInput,
		optFns ...func(*s3files.Options)) (*s3files.DeleteAccessPointOutput, error)
	ListMountTargets(ctx context.Context, params *s3files.ListMountTargetsInput,
		optFns ...func(*s3files.Options)) (*s3files.ListMountTargetsOutput, error)
	DeleteMountTarget(ctx context.Context, params *s3files.DeleteMountTargetInput,
		optFns ...func(*s3files.Options)) (*s3files.DeleteMountTargetOutput, error)
}
