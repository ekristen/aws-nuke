package resources

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/aws/aws-sdk-go-v2/service/s3files"
	s3filestypes "github.com/aws/aws-sdk-go-v2/service/s3files/types"

	"github.com/ekristen/aws-nuke/v3/mocks/mock_s3filesv2"
)

func Test_Mock_S3FilesFileSystem_List(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mock_s3filesv2.NewMockS3FilesAPI(ctrl)

	mockSvc.EXPECT().ListFileSystems(gomock.Any(), gomock.Any()).Return(&s3files.ListFileSystemsOutput{
		FileSystems: []s3filestypes.ListFileSystemsDescription{
			{FileSystemId: new("fs-1")},
			{FileSystemId: new("fs-2")},
		},
	}, nil)

	lister := &S3FilesFileSystemLister{svc: mockSvc}

	resources, err := lister.List(context.TODO(), testListerOpts)
	a.Nil(err)
	a.Len(resources, 2)
	a.Equal("fs-1", *resources[0].(*S3FilesFileSystem).ID)
	a.Equal("fs-2", *resources[1].(*S3FilesFileSystem).ID)
}

func Test_Mock_S3FilesFileSystem_List_Pagination(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mock_s3filesv2.NewMockS3FilesAPI(ctrl)

	gomock.InOrder(
		mockSvc.EXPECT().ListFileSystems(gomock.Any(), gomock.Any()).Return(&s3files.ListFileSystemsOutput{
			FileSystems: []s3filestypes.ListFileSystemsDescription{
				{FileSystemId: new("fs-1")},
			},
			NextToken: new("page-2"),
		}, nil),
		mockSvc.EXPECT().ListFileSystems(gomock.Any(), gomock.Any()).Return(&s3files.ListFileSystemsOutput{
			FileSystems: []s3filestypes.ListFileSystemsDescription{
				{FileSystemId: new("fs-2")},
			},
		}, nil),
	)

	lister := &S3FilesFileSystemLister{svc: mockSvc}

	resources, err := lister.List(context.TODO(), testListerOpts)
	a.Nil(err)
	a.Len(resources, 2)
}

func Test_Mock_S3FilesFileSystem_Remove(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mock_s3filesv2.NewMockS3FilesAPI(ctrl)

	r := &S3FilesFileSystem{
		svc: mockSvc,
		ID:  new("fs-1234567890abcdef"),
	}

	forceDelete := true
	mockSvc.EXPECT().DeleteFileSystem(gomock.Any(), gomock.Eq(&s3files.DeleteFileSystemInput{
		FileSystemId: r.ID,
		ForceDelete:  &forceDelete,
	})).Return(&s3files.DeleteFileSystemOutput{}, nil)

	err := r.Remove(context.TODO())
	a.Nil(err)
}
