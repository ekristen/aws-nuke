package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kinesisanalyticsv2"

	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const KinesisAnalyticsApplicationResource = "KinesisAnalyticsApplication"

func init() {
	resource.Register(resource.Registration{
		Name:   KinesisAnalyticsApplicationResource,
		Scope:  nuke.Account,
		Lister: &KinesisAnalyticsApplicationLister{},
	})
}

type KinesisAnalyticsApplicationLister struct{}

func (l *KinesisAnalyticsApplicationLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := kinesisanalyticsv2.New(opts.Session)

	resources := make([]resource.Resource, 0)

	params := &kinesisanalyticsv2.ListApplicationsInput{
		Limit: aws.Int64(25),
	}

	for {
		output, err := svc.ListApplications(params)
		if err != nil {
			return nil, err
		}

		for _, applicationSummary := range output.ApplicationSummaries {
			resources = append(resources, &KinesisAnalyticsApplication{
				svc:             svc,
				applicationName: applicationSummary.ApplicationName,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type KinesisAnalyticsApplication struct {
	svc             *kinesisanalyticsv2.KinesisAnalyticsV2
	applicationName *string
}

func (f *KinesisAnalyticsApplication) Remove(_ context.Context) error {
	output, err := f.svc.DescribeApplication(&kinesisanalyticsv2.DescribeApplicationInput{
		ApplicationName: f.applicationName,
	})

	if err != nil {
		return err
	}
	createTimestamp := output.ApplicationDetail.CreateTimestamp

	_, err = f.svc.DeleteApplication(&kinesisanalyticsv2.DeleteApplicationInput{
		ApplicationName: f.applicationName,
		CreateTimestamp: createTimestamp,
	})

	return err
}

func (f *KinesisAnalyticsApplication) String() string {
	return *f.applicationName
}
