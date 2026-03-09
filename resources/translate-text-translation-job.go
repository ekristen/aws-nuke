package resources

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/translate"
	translatetypes "github.com/aws/aws-sdk-go-v2/service/translate/types"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const TranslateTextTranslationJobResource = "TranslateTextTranslationJob"

func init() {
	registry.Register(&registry.Registration{
		Name:     TranslateTextTranslationJobResource,
		Scope:    nuke.Account,
		Resource: &TranslateTextTranslationJob{},
		Lister:   &TranslateTextTranslationJobLister{},
	})
}

type TranslateTextTranslationJobLister struct {
	svc TranslateAPI
}

func (l *TranslateTextTranslationJobLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	var resources []resource.Resource

	if l.svc == nil {
		opts := o.(*nuke.ListerOpts)
		l.svc = translate.NewFromConfig(*opts.Config)
	}

	params := &translate.ListTextTranslationJobsInput{
		MaxResults: aws.Int32(500),
	}

	for {
		resp, err := l.svc.ListTextTranslationJobs(ctx, params)
		if err != nil {
			return nil, err
		}

		for i := range resp.TextTranslationJobPropertiesList {
			item := &resp.TextTranslationJobPropertiesList[i]
			resources = append(resources, &TranslateTextTranslationJob{
				svc:                l.svc,
				JobID:              item.JobId,
				JobName:            item.JobName,
				JobStatus:          item.JobStatus,
				SourceLanguageCode: item.SourceLanguageCode,
				SubmittedTime:      item.SubmittedTime,
				EndTime:            item.EndTime,
			})
		}

		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return resources, nil
}

type TranslateTextTranslationJob struct {
	svc                TranslateAPI
	JobID              *string
	JobName            *string
	JobStatus          translatetypes.JobStatus
	SourceLanguageCode *string
	SubmittedTime      *time.Time
	EndTime            *time.Time
}

func (r *TranslateTextTranslationJob) Remove(ctx context.Context) error {
	_, err := r.svc.StopTextTranslationJob(ctx, &translate.StopTextTranslationJobInput{
		JobId: r.JobID,
	})
	return err
}

func (r *TranslateTextTranslationJob) Filter() error {
	switch r.JobStatus {
	case translatetypes.JobStatusSubmitted, translatetypes.JobStatusInProgress:
		return nil
	case translatetypes.JobStatusStopped, translatetypes.JobStatusStopRequested:
		return fmt.Errorf("translation job is already stopped or stop requested")
	case translatetypes.JobStatusCompleted, translatetypes.JobStatusCompletedWithError:
		return fmt.Errorf("translation job is already completed")
	case translatetypes.JobStatusFailed:
		return fmt.Errorf("translation job has failed")
	default:
		return fmt.Errorf("translation job status is %s, cannot be stopped", r.JobStatus)
	}
}

func (r *TranslateTextTranslationJob) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *TranslateTextTranslationJob) String() string {
	return *r.JobID
}
