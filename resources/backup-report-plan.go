package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/backup"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

type BackupReportPlan struct {
	svc  *backup.Backup
	arn  *string
	Name *string
}

type BackupReportPlanLister struct{}

func (BackupReportPlanLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
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
				svc:  svc,
				arn:  report.ReportPlanArn,
				Name: report.ReportPlanName,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

const BackupReportPlanResource = "BackupReportPlan"

func init() {
	registry.Register(&registry.Registration{
		Name:   BackupReportPlanResource,
		Scope:  nuke.Account,
		Lister: &BackupReportPlanLister{},
	})
}

func (r *BackupReportPlan) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("Name", r.Name)
	properties.Set("ARN", r.arn)
	return properties
}

func (r *BackupReportPlan) Remove(context.Context) error {
	_, err := r.svc.DeleteReportPlan(&backup.DeleteReportPlanInput{
		ReportPlanName: r.Name,
	})
	return err
}

func (r *BackupReportPlan) String() string {
	return *r.Name
}
