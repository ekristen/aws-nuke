package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/backup"
	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/rebuy-de/aws-nuke/v2/pkg/types"
)

type BackupReportPlan struct {
	svc            *backup.Backup
	arn            string
	reportPlanName string
}

type AWSBackupReportPlanLister struct{}

func (AWSBackupReportPlanLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := backup.New(opts.Session)
	max_backups_len := int64(100)
	params := &backup.ListReportPlansInput{
		MaxResults: &max_backups_len, // aws default limit on number of backup plans per account
	}
	resources := make([]resource.Resource, 0)

	for {
		output, err := svc.ListReportPlans(params)
		if err != nil {
			return nil, err
		}

		for _, report := range output.ReportPlans {
			resources = append(resources, &BackupReportPlan{
				svc:            svc,
				arn:            *report.ReportPlanArn,
				reportPlanName: *report.ReportPlanName,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

const AWSBackupReportPlanResource = "AWSBackupReportPlan"

func init() {
	registry.Register(&registry.Registration{
		Name:   AWSBackupReportPlanResource,
		Scope:  nuke.Account,
		Lister: &AWSBackupReportPlanLister{},
	})
}

func (b *BackupReportPlan) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("BackupReport", b.reportPlanName)
	return properties
}

func (b *BackupReportPlan) Remove(context.Context) error {
	_, err := b.svc.DeleteReportPlan(&backup.DeleteReportPlanInput{
		ReportPlanName: &b.reportPlanName,
	})
	return err
}

func (b *BackupReportPlan) String() string {
	return b.arn
}
