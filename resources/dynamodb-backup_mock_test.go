package resources

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"

	"github.com/aws/aws-sdk-go/service/dynamodb"

	"github.com/ekristen/aws-nuke/v3/mocks/mock_dynamodbiface"
)

func Test_Mock_DynamoDBBackup_List(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mock_dynamodbiface.NewMockDynamoDBAPI(ctrl)

	mockSvc.EXPECT().ListBackups(&dynamodb.ListBackupsInput{
		ExclusiveStartBackupArn: nil,
	}).Return(&dynamodb.ListBackupsOutput{
		BackupSummaries: []*dynamodb.BackupSummary{
			{
				BackupArn:              ptr.String("arn:aws:dynamodb:us-west-2:123456789012:table/ExampleTable/backup/1234567890123"),
				BackupName:             ptr.String("ExampleBackup"),
				BackupCreationDateTime: ptr.Time(time.Now()),
				TableName:              ptr.String("ExampleTable"),
				TableArn:               ptr.String("arn:aws:dynamodb:us-west-2:123456789012:table/ExampleTable"),
			},
		},
	}, nil)

	lister := &DynamoDBBackupLister{
		mockSvc: mockSvc,
	}

	resources, err := lister.List(context.TODO(), testListerOpts)
	a.Nil(err)
	a.Len(resources, 1)
}

func Test_Mock_DynamoDBBackup_Remove(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mock_dynamodbiface.NewMockDynamoDBAPI(ctrl)

	mockSvc.EXPECT().DeleteBackup(&dynamodb.DeleteBackupInput{
		BackupArn: ptr.String("arn:aws:dynamodb:us-west-2:123456789012:table/ExampleTable/backup/1234567890123"),
	}).Return(&dynamodb.DeleteBackupOutput{}, nil)

	now := time.Now()

	resource := &DynamoDBBackup{
		svc:        mockSvc,
		arn:        ptr.String("arn:aws:dynamodb:us-west-2:123456789012:table/ExampleTable/backup/1234567890123"),
		Name:       ptr.String("ExampleBackup"),
		CreateDate: ptr.Time(now),
		TableName:  ptr.String("ExampleTable"),
	}

	a.Equal(now.Format(time.RFC3339), resource.Properties().Get("CreateDate"))
	a.Equal("ExampleBackup", resource.Properties().Get("Name"))
	a.Equal("ExampleTable", resource.Properties().Get("TableName"))

	err := resource.Remove(context.TODO())
	a.Nil(err)
}
