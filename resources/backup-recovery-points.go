package resources

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/service/backup"           //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/backup/backupiface" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const AWSBackupRecoveryPointResource = "AWSBackupRecoveryPoint"

func init() {
	registry.Register(&registry.Registration{
		Name:     AWSBackupRecoveryPointResource,
		Scope:    nuke.Account,
		Resource: &BackupRecoveryPoint{},
		Lister:   &AWSBackupRecoveryPointLister{},
	})
}

type AWSBackupRecoveryPointLister struct {
	mockSvc backupiface.BackupAPI
}

func (l *AWSBackupRecoveryPointLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	var svc backupiface.BackupAPI
	if l.mockSvc != nil {
		svc = l.mockSvc
	} else {
		svc = backup.New(opts.Session)
	}

	maxVaultsLen := int64(100)
	params := &backup.ListBackupVaultsInput{
		MaxResults: &maxVaultsLen, // aws default limit on number of backup vaults per account
	}
	resp, err := svc.ListBackupVaults(params)
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, 0)
	for _, out := range resp.BackupVaultList {
		recoveryPointsOutput, _ := svc.ListRecoveryPointsByBackupVault(
			&backup.ListRecoveryPointsByBackupVaultInput{BackupVaultName: out.BackupVaultName})

		for _, rp := range recoveryPointsOutput.RecoveryPoints {
			tagsOutput, _ := svc.ListTags(&backup.ListTagsInput{ResourceArn: rp.RecoveryPointArn})
			resources = append(resources, &BackupRecoveryPoint{
				svc:             svc,
				arn:             *rp.RecoveryPointArn,
				backupVaultName: *out.BackupVaultName,
				tags:            tagsOutput.Tags,
			})
		}
	}

	return resources, nil
}

type BackupRecoveryPoint struct {
	svc             backupiface.BackupAPI
	arn             string
	backupVaultName string
	tags            map[string]*string
}

func (b *BackupRecoveryPoint) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("BackupVault", b.backupVaultName)
	for tagKey, tagValue := range b.tags {
		properties.Set(fmt.Sprintf("tag:%v", tagKey), *tagValue)
	}
	return properties
}

func (b *BackupRecoveryPoint) Remove(_ context.Context) error {
	_, err := b.svc.DeleteRecoveryPoint(&backup.DeleteRecoveryPointInput{
		BackupVaultName:  &b.backupVaultName,
		RecoveryPointArn: &b.arn,
	})
	return err
}

func (b *BackupRecoveryPoint) String() string {
	return b.arn
}
