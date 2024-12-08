package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/costandusagereportservice"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const BillingCostandUsageReportResource = "BillingCostandUsageReport"

func init() {
	registry.Register(&registry.Registration{
		Name:     BillingCostandUsageReportResource,
		Scope:    nuke.Account,
		Resource: &BillingCostandUsageReport{},
		Lister:   &BillingCostandUsageReportLister{},
	})
}

type BillingCostandUsageReportLister struct{}

type BillingCostandUsageReport struct {
	svc        *costandusagereportservice.CostandUsageReportService
	reportName *string
	s3Bucket   *string
	s3Prefix   *string
	s3Region   *string
}

func (l *BillingCostandUsageReportLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := costandusagereportservice.New(opts.Session)
	params := &costandusagereportservice.DescribeReportDefinitionsInput{
		MaxResults: aws.Int64(5),
	}

	reports := make([]*costandusagereportservice.ReportDefinition, 0)
	err := svc.DescribeReportDefinitionsPages(
		params, func(page *costandusagereportservice.DescribeReportDefinitionsOutput, lastPage bool) bool {
			reports = append(reports, page.ReportDefinitions...)
			return true
		})
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, 0)
	for _, report := range reports {
		resources = append(resources, &BillingCostandUsageReport{
			svc:        svc,
			reportName: report.ReportName,
			s3Bucket:   report.S3Bucket,
			s3Prefix:   report.S3Prefix,
			s3Region:   report.S3Region,
		})
	}

	return resources, nil
}

func (r *BillingCostandUsageReport) Remove(_ context.Context) error {
	_, err := r.svc.DeleteReportDefinition(&costandusagereportservice.DeleteReportDefinitionInput{
		ReportName: r.reportName,
	})

	return err
}

func (r *BillingCostandUsageReport) Properties() types.Properties {
	properties := types.NewProperties()
	properties.
		Set("Name", *r.reportName).
		Set("S3Bucket", *r.s3Bucket).
		Set("s3Prefix", *r.s3Prefix).
		Set("S3Region", *r.s3Region)
	return properties
}

func (r *BillingCostandUsageReport) String() string {
	return *r.reportName
}
