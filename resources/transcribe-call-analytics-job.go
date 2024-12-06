package resources

import (
	"context"
	"errors"
	"strings"
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
		Name:     TranscribeCallAnalyticsJobResource,
		Scope:    nuke.Account,
		Resource: &TranscribeCallAnalyticsJob{},
		Lister:   &TranscribeCallAnalyticsJobLister{},
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
			var badRequestException *transcribeservice.BadRequestException
			if errors.As(err, &badRequestException) &&
				strings.Contains(badRequestException.Message(), "isn't supported in this region") {
				return resources, nil
			}
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
	properties.Set("CompletionTime", r.completionTime)
	properties.Set("CreationTime", r.creationTime)
	properties.Set("FailureReason", r.failureReason)
	properties.Set("LanguageCode", r.languageCode)
	properties.Set("StartTime", r.startTime)
	return properties
}

func (r *TranscribeCallAnalyticsJob) String() string {
	return *r.name
}
