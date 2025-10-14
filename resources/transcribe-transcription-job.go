package resources

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/aws"                       //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/transcribeservice" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const TranscribeTranscriptionJobResource = "TranscribeTranscriptionJob"

func init() {
	registry.Register(&registry.Registration{
		Name:     TranscribeTranscriptionJobResource,
		Scope:    nuke.Account,
		Resource: &TranscribeTranscriptionJob{},
		Lister:   &TranscribeTranscriptionJobLister{},
	})
}

type TranscribeTranscriptionJobLister struct{}

func (l *TranscribeTranscriptionJobLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := transcribeservice.New(opts.Session)
	resources := make([]resource.Resource, 0)
	var nextToken *string

	for {
		listTranscriptionJobsInput := &transcribeservice.ListTranscriptionJobsInput{
			MaxResults: aws.Int64(100),
			NextToken:  nextToken,
		}

		listOutput, err := svc.ListTranscriptionJobs(listTranscriptionJobsInput)
		if err != nil {
			return nil, err
		}
		for _, job := range listOutput.TranscriptionJobSummaries {
			resources = append(resources, &TranscribeTranscriptionJob{
				svc:            svc,
				name:           job.TranscriptionJobName,
				status:         job.TranscriptionJobStatus,
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

type TranscribeTranscriptionJob struct {
	svc            *transcribeservice.TranscribeService
	name           *string
	status         *string
	completionTime *time.Time
	creationTime   *time.Time
	failureReason  *string
	languageCode   *string
	startTime      *time.Time
}

func (r *TranscribeTranscriptionJob) Remove(_ context.Context) error {
	deleteInput := &transcribeservice.DeleteTranscriptionJobInput{
		TranscriptionJobName: r.name,
	}
	_, err := r.svc.DeleteTranscriptionJob(deleteInput)
	return err
}

func (r *TranscribeTranscriptionJob) Properties() types.Properties {
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

func (r *TranscribeTranscriptionJob) String() string {
	return *r.name
}
