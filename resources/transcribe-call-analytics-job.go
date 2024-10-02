package resources

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/transcribeservice"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const TranscribeCallAnalyticsJobResource = "TranscribeCallAnalyticsJob"

func init() {
	registry.Register(&registry.Registration{
		Name:   TranscribeCallAnalyticsJobResource,
		Scope:  nuke.Account,
		Lister: &TranscribeCallAnalyticsJobLister{},
	})
}

type TranscribeCallAnalyticsJobLister struct{}

func (l *TranscribeCallAnalyticsJobLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := transcribeservice.New(opts.Session)
	resources := make([]resource.Resource, 0)
	var nextToken *string

	for {
		listCallAnalyticsJobsInput := &transcribeservice.ListCallAnalyticsJobsInput{
			MaxResults: aws.Int64(100),
			NextToken:  nextToken,
		}

		listOutput, err := svc.ListCallAnalyticsJobs(listCallAnalyticsJobsInput)
		if err != nil {
			return nil, err
		}
		for _, job := range listOutput.CallAnalyticsJobSummaries {
			resources = append(resources, &TranscribeCallAnalyticsJob{
				svc:            svc,
				name:           job.CallAnalyticsJobName,
				status:         job.CallAnalyticsJobStatus,
				completionTime: job.CompletionTime,
				creationTime:   job.CreationTime,
				failureReason:  job.FailureReason,
				languageCode:   job.LanguageCode,
				startTime:      job.StartTime,
			})
		}

		// Check if there are more results
		if listOutput.NextToken == nil {
			break // No more results, exit the loop
		}

		// Set the nextToken for the next iteration
		nextToken = listOutput.NextToken
	}
	return resources, nil
}

type TranscribeCallAnalyticsJob struct {
	svc            *transcribeservice.TranscribeService
	name           *string
	status         *string
	completionTime *time.Time
	creationTime   *time.Time
	failureReason  *string
	languageCode   *string
	startTime      *time.Time
}

func (r *TranscribeCallAnalyticsJob) Remove(_ context.Context) error {
	deleteInput := &transcribeservice.DeleteCallAnalyticsJobInput{
		CallAnalyticsJobName: r.name,
	}
	_, err := r.svc.DeleteCallAnalyticsJob(deleteInput)
	return err
}

func (r *TranscribeCallAnalyticsJob) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("Name", r.name)
	properties.Set("Status", r.status)
	if r.completionTime != nil {
		properties.Set("CompletionTime", r.completionTime.Format(time.RFC3339))
	}
	if r.creationTime != nil {
		properties.Set("CreationTime", r.creationTime.Format(time.RFC3339))
	}
	properties.Set("FailureReason", r.failureReason)
	properties.Set("LanguageCode", r.languageCode)
	if r.startTime != nil {
		properties.Set("StartTime", r.startTime.Format(time.RFC3339))
	}
	return properties
}

func (r *TranscribeCallAnalyticsJob) String() string {
	return *r.name
}