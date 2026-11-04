//go:generate ../mocks/generate_mocks.sh backup backupiface
package resources

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"

	"github.com/aws/aws-sdk-go/service/backup" //nolint:staticcheck

	"github.com/ekristen/aws-nuke/v3/mocks/mock_backupiface"
)

func Test_Mock_AWSBackupRecoveryPoint_List(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mock_backupiface.NewMockBackupAPI(ctrl)

	mockSvc.EXPECT().
		ListBackupVaults(gomock.Any()).
		Return(&backup.ListBackupVaultsOutput{
			BackupVaultList: []*backup.VaultListMember{
				{
					BackupVaultName: ptr.String("test-vault"),
					BackupVaultArn:  ptr.String("arn:aws:backup:us-east-1:123456789012:backup-vault:test-vault"),
				},
			},
		}, nil)

	mockSvc.EXPECT().
		ListRecoveryPointsByBackupVault(gomock.Eq(&backup.ListRecoveryPointsByBackupVaultInput{
			BackupVaultName: ptr.String("test-vault"),
		})).
		Return(&backup.ListRecoveryPointsByBackupVaultOutput{
			RecoveryPoints: []*backup.RecoveryPointByBackupVault{
				{
					RecoveryPointArn: ptr.String("arn:aws:backup:us-east-1:123456789012:recovery-point:test-rp"),
				},
			},
		}, nil)

	mockSvc.EXPECT().
		ListTags(gomock.Eq(&backup.ListTagsInput{
			ResourceArn: ptr.String("arn:aws:backup:us-east-1:123456789012:recovery-point:test-rp"),
		})).
		Return(&backup.ListTagsOutput{
			Tags: map[string]*string{
				"Environment": ptr.String("test"),
			},
		}, nil)

	lister := AWSBackupRecoveryPointLister{
		mockSvc: mockSvc,
	}

	resources, err := lister.List(context.TODO(), testListerOpts)

	a.Nil(err)
	a.Len(resources, 1)
}

func Test_Mock_AWSBackupRecoveryPoint_Remove(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mock_backupiface.NewMockBackupAPI(ctrl)

	mockSvc.EXPECT().
		DeleteRecoveryPoint(gomock.Eq(&backup.DeleteRecoveryPointInput{
			BackupVaultName:  ptr.String("test-vault"),
			RecoveryPointArn: ptr.String("arn:aws:backup:us-east-1:123456789012:recovery-point:test-rp"),
		})).
		Return(&backup.DeleteRecoveryPointOutput{}, nil)

	resource := BackupRecoveryPoint{
		svc:             mockSvc,
		arn:             "arn:aws:backup:us-east-1:123456789012:recovery-point:test-rp",
		backupVaultName: "test-vault",
	}

	err := resource.Remove(context.TODO())
	a.Nil(err)
}

func Test_Mock_AWSBackupRecoveryPoint_Properties(t *testing.T) {
	a := assert.New(t)

	resource := BackupRecoveryPoint{
		arn:             "arn:aws:backup:us-east-1:123456789012:recovery-point:test-rp",
		backupVaultName: "test-vault",
		tags: map[string]*string{
			"Environment": ptr.String("test"),
		},
	}

	properties := resource.Properties()

	a.Equal("test-vault", properties.Get("BackupVault"))
	a.Equal("test", properties.Get("tag:Environment"))
}
