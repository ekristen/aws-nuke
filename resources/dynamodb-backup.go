package resources

import (
	"context"
	"time"

	"github.com/gotidy/ptr"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const DynamoDBBackupResource = "DynamoDBBackup"

func init() {
	registry.Register(&registry.Registration{
		Name:     DynamoDBBackupResource,
		Scope:    nuke.Account,
		Resource: &DynamoDBBackup{},
		Lister:   &DynamoDBBackupLister{},
	})
}

type DynamoDBBackupLister struct {
	mockSvc dynamodbiface.DynamoDBAPI
}

func (l *DynamoDBBackupLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	var svc dynamodbiface.DynamoDBAPI
	if l.mockSvc != nil {
		svc = l.mockSvc
	} else {
		svc = dynamodb.New(opts.Session)
	}

	resources := make([]resource.Resource, 0)

	var lastEvaluatedBackupArn *string

	for {
		backupsResp, err := svc.ListBackups(&dynamodb.ListBackupsInput{
			ExclusiveStartBackupArn: lastEvaluatedBackupArn,
		})
		if err != nil {
			return nil, err
		}

		for _, backup := range backupsResp.BackupSummaries {
			resources = append(resources, &DynamoDBBackup{
				svc:        svc,
				arn:        backup.BackupArn,
				Name:       backup.BackupName,
				CreateDate: backup.BackupCreationDateTime,
				TableName:  backup.TableName,
			})
		}

		if backupsResp.LastEvaluatedBackupArn == nil {
			break
		}

		lastEvaluatedBackupArn = backupsResp.LastEvaluatedBackupArn
	}

	return resources, nil
}

type DynamoDBBackup struct {
	svc        dynamodbiface.DynamoDBAPI
	arn        *string
	Name       *string
	CreateDate *time.Time
	TableName  *string
}

func (r *DynamoDBBackup) Remove(_ context.Context) error {
	params := &dynamodb.DeleteBackupInput{
		BackupArn: r.arn,
	}

	_, err := r.svc.DeleteBackup(params)
	if err != nil {
		return err
	}

	return nil
}

func (r *DynamoDBBackup) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *DynamoDBBackup) String() string {
	return ptr.ToString(r.Name)
}
