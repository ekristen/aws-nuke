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

const TranscribeMedicalTranscriptionJobResource = "TranscribeMedicalTranscriptionJob"

func init() {
	registry.Register(&registry.Registration{
		Name:     TranscribeMedicalTranscriptionJobResource,
		Scope:    nuke.Account,
		Resource: &TranscribeMedicalTranscriptionJob{},
		Lister:   &TranscribeMedicalTranscriptionJobLister{},
	})
}

type TranscribeMedicalTranscriptionJobLister struct{}

func (l *TranscribeMedicalTranscriptionJobLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := transcribeservice.New(opts.Session)
	resources := make([]resource.Resource, 0)
	var nextToken *string

	for {
		listMedicalTranscriptionJobsInput := &transcribeservice.ListMedicalTranscriptionJobsInput{
			MaxResults: aws.Int64(100),
			NextToken:  nextToken,
		}

		listOutput, err := svc.ListMedicalTranscriptionJobs(listMedicalTranscriptionJobsInput)
		if err != nil {
			return nil, err
		}
		for _, job := range listOutput.MedicalTranscriptionJobSummaries {
			resources = append(resources, &TranscribeMedicalTranscriptionJob{
				svc:                       svc,
				name:                      job.MedicalTranscriptionJobName,
				status:                    job.TranscriptionJobStatus,
				completionTime:            job.CompletionTime,
				contentIdentificationType: job.ContentIdentificationType,
				creationTime:              job.CreationTime,
				failureReason:             job.FailureReason,
				languageCode:              job.LanguageCode,
				outputLocationType:        job.OutputLocationType,
				specialty:                 job.Specialty,
				startTime:                 job.StartTime,
				inputType:                 job.Type,
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

type TranscribeMedicalTranscriptionJob struct {
	svc                       *transcribeservice.TranscribeService
	name                      *string
	status                    *string
	completionTime            *time.Time
	contentIdentificationType *string
	creationTime              *time.Time
	failureReason             *string
	languageCode              *string
	outputLocationType        *string
	specialty                 *string
	startTime                 *time.Time
	inputType                 *string
}

func (r *TranscribeMedicalTranscriptionJob) Remove(_ context.Context) error {
	deleteInput := &transcribeservice.DeleteMedicalTranscriptionJobInput{
		MedicalTranscriptionJobName: r.name,
	}
	_, err := r.svc.DeleteMedicalTranscriptionJob(deleteInput)
	return err
}

func (r *TranscribeMedicalTranscriptionJob) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("Name", r.name)
	properties.Set("Status", r.status)
	properties.Set("CompletionTime", r.completionTime)
	properties.Set("ContentIdentificationType", r.contentIdentificationType)
	properties.Set("CreationTime", r.creationTime)
	properties.Set("FailureReason", r.failureReason)
	properties.Set("LanguageCode", r.languageCode)
	properties.Set("OutputLocationType", r.outputLocationType)
	properties.Set("Specialty", r.specialty)
	properties.Set("StartTime", r.startTime)
	properties.Set("InputType", r.inputType)
	return properties
}

func (r *TranscribeMedicalTranscriptionJob) String() string {
	return *r.name
}
