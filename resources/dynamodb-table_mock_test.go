package resources

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"

	"github.com/aws/aws-sdk-go/aws/awserr"       //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/dynamodb" //nolint:staticcheck

	libsettings "github.com/ekristen/libnuke/pkg/settings"

	"github.com/ekristen/aws-nuke/v3/mocks/mock_dynamodbiface"
)

func Test_Mock_DynamoDBTable_List(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mock_dynamodbiface.NewMockDynamoDBAPI(ctrl)

	mockSvc.EXPECT().ListTables(&dynamodb.ListTablesInput{}).Return(&dynamodb.ListTablesOutput{
		TableNames: []*string{
			ptr.String("ExampleTable"),
		},
	}, nil)

	mockSvc.EXPECT().DescribeTable(&dynamodb.DescribeTableInput{
		TableName: ptr.String("ExampleTable"),
	}).Return(&dynamodb.DescribeTableOutput{
		Table: &dynamodb.TableDescription{
			TableArn: ptr.String("arn:aws:dynamodb:us-west-2:123456789012:table/ExampleTable"),
		},
	}, nil)

	mockSvc.EXPECT().ListTagsOfResource(&dynamodb.ListTagsOfResourceInput{
		ResourceArn: ptr.String("arn:aws:dynamodb:us-west-2:123456789012:table/ExampleTable"),
	}).Return(&dynamodb.ListTagsOfResourceOutput{
		Tags: []*dynamodb.Tag{
			{
				Key:   ptr.String("Name"),
				Value: ptr.String("ExampleTable"),
			},
		},
	}, nil)

	lister := &DynamoDBTableLister{
		mockSvc: mockSvc,
	}

	resources, err := lister.List(context.TODO(), testListerOpts)
	a.NoError(err)
	a.Len(resources, 1)
}

func Test_Mock_DynamoDBTable_List_Paginated(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mock_dynamodbiface.NewMockDynamoDBAPI(ctrl)

	gomock.InOrder(
		mockSvc.EXPECT().ListTables(&dynamodb.ListTablesInput{}).Return(&dynamodb.ListTablesOutput{
			TableNames:             []*string{ptr.String("TableA")},
			LastEvaluatedTableName: ptr.String("TableA"),
		}, nil),
		mockSvc.EXPECT().ListTables(&dynamodb.ListTablesInput{
			ExclusiveStartTableName: ptr.String("TableA"),
		}).Return(&dynamodb.ListTablesOutput{
			TableNames: []*string{ptr.String("TableB")},
		}, nil),
	)

	for _, name := range []string{"TableA", "TableB"} {
		arn := "arn:aws:dynamodb:us-west-2:123456789012:table/" + name
		mockSvc.EXPECT().DescribeTable(&dynamodb.DescribeTableInput{
			TableName: ptr.String(name),
		}).Return(&dynamodb.DescribeTableOutput{
			Table: &dynamodb.TableDescription{
				TableArn: ptr.String(arn),
			},
		}, nil)

		mockSvc.EXPECT().ListTagsOfResource(&dynamodb.ListTagsOfResourceInput{
			ResourceArn: ptr.String(arn),
		}).Return(&dynamodb.ListTagsOfResourceOutput{}, nil)
	}

	lister := &DynamoDBTableLister{
		mockSvc: mockSvc,
	}

	resources, err := lister.List(context.TODO(), testListerOpts)
	a.NoError(err)
	a.Len(resources, 2)
	a.Equal("TableA", resources[0].(*DynamoDBTable).String())
	a.Equal("TableB", resources[1].(*DynamoDBTable).String())
}

func Test_Mock_DynamoDBTable_Remove(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mock_dynamodbiface.NewMockDynamoDBAPI(ctrl)

	mockSvc.EXPECT().DeleteTable(&dynamodb.DeleteTableInput{
		TableName: ptr.String("ExampleTable"),
	})

	settings := &libsettings.Setting{}
	settings.Set("DisableDeletionProtection", false)

	resource := &DynamoDBTable{
		svc:        mockSvc,
		settings:   settings,
		id:         ptr.String("ExampleTable"),
		protection: ptr.Bool(false),
		Name:       ptr.String("ExampleTable"),
		Tags: []*dynamodb.Tag{
			{
				Key:   ptr.String("Name"),
				Value: ptr.String("ExampleTable"),
			},
		},
	}

	err := resource.Remove(context.TODO())
	a.Nil(err)
}

func Test_Mock_DynamoDBTable_Remove_DeletionProtection(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mock_dynamodbiface.NewMockDynamoDBAPI(ctrl)

	mockSvc.EXPECT().DeleteTable(&dynamodb.DeleteTableInput{
		TableName: ptr.String("ExampleTable"),
	}).Return(nil,
		awserr.New("ValidationException",
			"Resource cannot be deleted as it is currently protected against deletion. "+
				"Disable deletion protection first.", nil))

	settings := &libsettings.Setting{}
	settings.Set("DisableDeletionProtection", false)

	resource := &DynamoDBTable{
		svc:        mockSvc,
		settings:   settings,
		id:         ptr.String("ExampleTable"),
		protection: ptr.Bool(true),
		Name:       ptr.String("ExampleTable"),
		Tags: []*dynamodb.Tag{
			{
				Key:   ptr.String("Name"),
				Value: ptr.String("ExampleTable"),
			},
		},
	}

	err := resource.Remove(context.TODO())
	a.Error(err)
}

func Test_Mock_DynamoDBTable_Remove_DeletionProtection_Disable(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mock_dynamodbiface.NewMockDynamoDBAPI(ctrl)

	mockSvc.EXPECT().UpdateTable(&dynamodb.UpdateTableInput{
		TableName:                 ptr.String("ExampleTable"),
		DeletionProtectionEnabled: ptr.Bool(false),
	}).Return(&dynamodb.UpdateTableOutput{}, nil)

	mockSvc.EXPECT().DeleteTable(&dynamodb.DeleteTableInput{
		TableName: ptr.String("ExampleTable"),
	}).Return(nil,
		awserr.New("ValidationException",
			"Resource cannot be deleted as it is currently protected against deletion. "+
				"Disable deletion protection first.", nil))

	settings := &libsettings.Setting{}
	settings.Set("DisableDeletionProtection", true)

	resource := &DynamoDBTable{
		svc:        mockSvc,
		settings:   settings,
		id:         ptr.String("ExampleTable"),
		protection: ptr.Bool(true),
		Name:       ptr.String("ExampleTable"),
		Tags: []*dynamodb.Tag{
			{
				Key:   ptr.String("Name"),
				Value: ptr.String("ExampleTable"),
			},
		},
	}

	err := resource.Remove(context.TODO())
	a.Error(err)
}
