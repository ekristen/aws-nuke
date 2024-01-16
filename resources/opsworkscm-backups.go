package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/opsworkscm"

	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const OpsWorksCMBackupResource = "OpsWorksCMBackup"

func init() {
	resource.Register(resource.Registration{
		Name:   OpsWorksCMBackupResource,
		Scope:  nuke.Account,
		Lister: &OpsWorksCMBackupLister{},
	})
}

type OpsWorksCMBackupLister struct{}

func (l *OpsWorksCMBackupLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := opsworkscm.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &opsworkscm.DescribeBackupsInput{}

	output, err := svc.DescribeBackups(params)
	if err != nil {
		return nil, err
	}

	for _, backup := range output.Backups {
		resources = append(resources, &OpsWorksCMBackup{
			svc: svc,
			ID:  backup.BackupId,
		})
	}

	return resources, nil
}

type OpsWorksCMBackup struct {
	svc *opsworkscm.OpsWorksCM
	ID  *string
}

func (f *OpsWorksCMBackup) Remove(_ context.Context) error {
	_, err := f.svc.DeleteBackup(&opsworkscm.DeleteBackupInput{
		BackupId: f.ID,
	})

	return err
}

func (f *OpsWorksCMBackup) String() string {
	return *f.ID
}
