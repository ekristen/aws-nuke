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

func Test_Mock_S3FilesMountTarget_List(t *testing.T) {
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

	mockSvc.EXPECT().ListMountTargets(gomock.Any(), gomock.Eq(&s3files.ListMountTargetsInput{
		FileSystemId: new("fs-1"),
	})).Return(&s3files.ListMountTargetsOutput{
		MountTargets: []s3filestypes.ListMountTargetsDescription{
			{MountTargetId: new("fsmt-1")},
			{MountTargetId: new("fsmt-2")},
		},
	}, nil)

	mockSvc.EXPECT().ListMountTargets(gomock.Any(), gomock.Eq(&s3files.ListMountTargetsInput{
		FileSystemId: new("fs-2"),
	})).Return(&s3files.ListMountTargetsOutput{
		MountTargets: []s3filestypes.ListMountTargetsDescription{},
	}, nil)

	lister := &S3FilesMountTargetLister{svc: mockSvc}

	resources, err := lister.List(context.TODO(), testListerOpts)
	a.Nil(err)
	a.Len(resources, 2)
	a.Equal("fsmt-1", *resources[0].(*S3FilesMountTarget).ID)
	a.Equal("fsmt-2", *resources[1].(*S3FilesMountTarget).ID)
}

func Test_Mock_S3FilesMountTarget_Remove(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mock_s3filesv2.NewMockS3FilesAPI(ctrl)

	r := &S3FilesMountTarget{
		svc: mockSvc,
		ID:  new("fsmt-1234567890abcdef"),
	}

	mockSvc.EXPECT().DeleteMountTarget(gomock.Any(), gomock.Eq(&s3files.DeleteMountTargetInput{
		MountTargetId: r.ID,
	})).Return(&s3files.DeleteMountTargetOutput{}, nil)

	err := r.Remove(context.TODO())
	a.Nil(err)
}
