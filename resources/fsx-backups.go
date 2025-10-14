package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"         //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/fsx" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const FSxBackupResource = "FSxBackup"

func init() {
	registry.Register(&registry.Registration{
		Name:     FSxBackupResource,
		Scope:    nuke.Account,
		Resource: &FSxBackup{},
		Lister:   &FSxBackupLister{},
	})
}

type FSxBackupLister struct{}

func (l *FSxBackupLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := fsx.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &fsx.DescribeBackupsInput{
		MaxResults: aws.Int64(100),
	}

	for {
		resp, err := svc.DescribeBackups(params)
		if err != nil {
			return nil, err
		}

		for _, backup := range resp.Backups {
			resources = append(resources, &FSxBackup{
				svc:    svc,
				backup: backup,
			})
		}
		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}
	return resources, nil
}

type FSxBackup struct {
	svc    *fsx.FSx
	backup *fsx.Backup
}

func (f *FSxBackup) Remove(_ context.Context) error {
	_, err := f.svc.DeleteBackup(&fsx.DeleteBackupInput{
		BackupId: f.backup.BackupId,
	})

	return err
}

func (f *FSxBackup) Properties() types.Properties {
	properties := types.NewProperties()
	for _, tagValue := range f.backup.Tags {
		properties.SetTag(tagValue.Key, tagValue.Value)
	}
	properties.Set("Type", f.backup.Type)
	return properties
}

func (f *FSxBackup) String() string {
	return *f.backup.BackupId
}
