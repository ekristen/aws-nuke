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

func Test_Mock_S3FilesAccessPoint_List(t *testing.T) {
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

	mockSvc.EXPECT().ListAccessPoints(gomock.Any(), gomock.Eq(&s3files.ListAccessPointsInput{
		FileSystemId: new("fs-1"),
	})).Return(&s3files.ListAccessPointsOutput{
		AccessPoints: []s3filestypes.ListAccessPointsDescription{
			{AccessPointId: new("fsap-1a")},
			{AccessPointId: new("fsap-1b")},
		},
	}, nil)

	mockSvc.EXPECT().ListAccessPoints(gomock.Any(), gomock.Eq(&s3files.ListAccessPointsInput{
		FileSystemId: new("fs-2"),
	})).Return(&s3files.ListAccessPointsOutput{
		AccessPoints: []s3filestypes.ListAccessPointsDescription{
			{AccessPointId: new("fsap-2a")},
		},
	}, nil)

	lister := &S3FilesAccessPointLister{svc: mockSvc}

	resources, err := lister.List(context.TODO(), testListerOpts)
	a.Nil(err)
	a.Len(resources, 3)
	a.Equal("fsap-1a", *resources[0].(*S3FilesAccessPoint).ID)
	a.Equal("fsap-1b", *resources[1].(*S3FilesAccessPoint).ID)
	a.Equal("fsap-2a", *resources[2].(*S3FilesAccessPoint).ID)
}

func Test_Mock_S3FilesAccessPoint_Remove(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mock_s3filesv2.NewMockS3FilesAPI(ctrl)

	r := &S3FilesAccessPoint{
		svc: mockSvc,
		ID:  new("fsap-1234567890abcdef"),
	}

	mockSvc.EXPECT().DeleteAccessPoint(gomock.Any(), gomock.Eq(&s3files.DeleteAccessPointInput{
		AccessPointId: r.ID,
	})).Return(&s3files.DeleteAccessPointOutput{}, nil)

	err := r.Remove(context.TODO())
	a.Nil(err)
}
