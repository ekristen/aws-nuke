package resources

import (
	"context"

	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/service/backup"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

type BackupSelection struct {
	svc           *backup.Backup
	planID        string
	selectionID   string
	selectionName string
}

const AWSBackupSelectionResource = "AWSBackupSelection"

func init() {
	registry.Register(&registry.Registration{
		Name:   AWSBackupSelectionResource,
		Scope:  nuke.Account,
		Lister: &AWSBackupSelectionLister{},
	})
}

type AWSBackupSelectionLister struct{}

func (l *AWSBackupSelectionLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := backup.New(opts.Session)
	falseValue := false
	maxBackupsLen := int64(100)
	params := &backup.ListBackupPlansInput{
		IncludeDeleted: &falseValue,
		MaxResults:     &maxBackupsLen, // aws default limit on number of backup plans per account
	}
	resources := make([]resource.Resource, 0)

	for {
		output, err := svc.ListBackupPlans(params)
		if err != nil {
			return nil, err
		}

		for _, plan := range output.BackupPlansList {
			selectionsOutput, _ := svc.ListBackupSelections(&backup.ListBackupSelectionsInput{BackupPlanId: plan.BackupPlanId})
			for _, selection := range selectionsOutput.BackupSelectionsList {
				resources = append(resources, &BackupSelection{
					svc:           svc,
					planID:        *selection.BackupPlanId,
					selectionID:   *selection.SelectionId,
					selectionName: *selection.SelectionName,
				})
			}
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

func (b *BackupSelection) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("Name", b.selectionName)
	properties.Set("ID", b.selectionID)
	properties.Set("PlanID", b.planID)
	return properties
}

func (b *BackupSelection) Remove(_ context.Context) error {
	_, err := b.svc.DeleteBackupSelection(&backup.DeleteBackupSelectionInput{
		BackupPlanId: &b.planID,
		SelectionId:  &b.selectionID,
	})
	return err
}

func (b *BackupSelection) String() string {
	return fmt.Sprintf("%s (%s)", b.planID, b.selectionID)
}

func (b *BackupSelection) Filter() error {
	if strings.HasPrefix(b.selectionName, "aws/efs/") {
		return fmt.Errorf("cannot delete EFS automatic backups backup selection")
	}
	return nil
}
