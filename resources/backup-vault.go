package resources

import (
	"context"

	"fmt"

	"github.com/aws/aws-sdk-go/service/backup"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

type BackupVault struct {
	svc  *backup.Backup
	arn  string
	name string
	tags map[string]*string
}

const BackupVaultResource = "BackupVault"

func init() {
	registry.Register(&registry.Registration{
		Name:   BackupVaultResource,
		Scope:  nuke.Account,
		Lister: &AWSBackupVaultLister{},
		DeprecatedAliases: []string{
			"AWSBackupVault",
		},
	})
}

type AWSBackupVaultLister struct{}

func (l *AWSBackupVaultLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := backup.New(opts.Session)
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
		tagsOutput, _ := svc.ListTags(&backup.ListTagsInput{ResourceArn: out.BackupVaultArn})
		resources = append(resources, &BackupVault{
			svc:  svc,
			name: *out.BackupVaultName,
			arn:  *out.BackupVaultArn,
			tags: tagsOutput.Tags,
		})
	}

	return resources, nil
}

func (b *BackupVault) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("Name", b.name)
	for tagKey, tagValue := range b.tags {
		properties.Set(fmt.Sprintf("tag:%v", tagKey), *tagValue)
	}
	return properties
}

func (b *BackupVault) Remove(_ context.Context) error {
	_, err := b.svc.DeleteBackupVault(&backup.DeleteBackupVaultInput{
		BackupVaultName: &b.name,
	})
	return err
}

func (b *BackupVault) String() string {
	return b.arn
}

func (b *BackupVault) Filter() error {
	if b.name == "aws/efs/automatic-backup-vault" {
		return fmt.Errorf("cannot delete EFS automatic backups vault")
	}
	return nil
}
