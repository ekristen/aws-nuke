package resources

import (
	"context"
	"testing"
	"time"

	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/aws/aws-sdk-go-v2/service/translate"
	translatetypes "github.com/aws/aws-sdk-go-v2/service/translate/types"

	"github.com/ekristen/aws-nuke/v3/mocks/mock_translate"
)

func Test_Mock_TranslateTextTranslationJob_List(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mock_translate.NewMockTranslateAPI(ctrl)

	now := time.Now().UTC()

	mockSvc.EXPECT().ListTextTranslationJobs(gomock.Any(), gomock.Any()).Return(&translate.ListTextTranslationJobsOutput{
		TextTranslationJobPropertiesList: []translatetypes.TextTranslationJobProperties{
			{
				JobId:              ptr.String("job-123"),
				JobName:            ptr.String("my-translation-job"),
				JobStatus:          translatetypes.JobStatusInProgress,
				SourceLanguageCode: ptr.String("en"),
				SubmittedTime:      &now,
			},
			{
				JobId:              ptr.String("job-456"),
				JobName:            ptr.String("completed-job"),
				JobStatus:          translatetypes.JobStatusCompleted,
				SourceLanguageCode: ptr.String("en"),
				SubmittedTime:      &now,
				EndTime:            &now,
			},
		},
	}, nil)

	lister := &TranslateTextTranslationJobLister{
		svc: mockSvc,
	}

	resources, err := lister.List(context.TODO(), testListerOpts)
	a.Nil(err)
	a.Len(resources, 2)

	resource := resources[0].(*TranslateTextTranslationJob)
	a.Equal("job-123", *resource.JobID)
	a.Equal(translatetypes.JobStatusInProgress, resource.JobStatus)
}

func Test_Mock_TranslateTextTranslationJob_Filter(t *testing.T) {
	a := assert.New(t)

	cases := []struct {
		Status   translatetypes.JobStatus
		Filtered bool
	}{
		{Status: translatetypes.JobStatusSubmitted, Filtered: false},
		{Status: translatetypes.JobStatusInProgress, Filtered: false},
		{Status: translatetypes.JobStatusCompleted, Filtered: true},
		{Status: translatetypes.JobStatusCompletedWithError, Filtered: true},
		{Status: translatetypes.JobStatusFailed, Filtered: true},
		{Status: translatetypes.JobStatusStopRequested, Filtered: true},
		{Status: translatetypes.JobStatusStopped, Filtered: true},
	}

	for _, c := range cases {
		t.Run(string(c.Status), func(t *testing.T) {
			r := &TranslateTextTranslationJob{
				JobID:     ptr.String("job-123"),
				JobStatus: c.Status,
			}

			err := r.Filter()
			if c.Filtered {
				a.NotNil(err)
			} else {
				a.Nil(err)
			}
		})
	}
}

func Test_Mock_TranslateTextTranslationJob_Remove(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mock_translate.NewMockTranslateAPI(ctrl)

	mockSvc.EXPECT().
		StopTextTranslationJob(gomock.Any(), gomock.Any()).
		Return(&translate.StopTextTranslationJobOutput{}, nil)

	r := &TranslateTextTranslationJob{
		svc:   mockSvc,
		JobID: ptr.String("job-123"),
	}

	err := r.Remove(context.TODO())
	a.Nil(err)
}

func Test_Mock_TranslateTextTranslationJob_Properties(t *testing.T) {
	a := assert.New(t)

	now := time.Now().UTC()

	r := &TranslateTextTranslationJob{
		JobID:              ptr.String("job-123"),
		JobName:            ptr.String("my-translation-job"),
		JobStatus:          translatetypes.JobStatusInProgress,
		SourceLanguageCode: ptr.String("en"),
		SubmittedTime:      &now,
	}

	props := r.Properties()
	a.Equal("job-123", props.Get("JobID"))
	a.Equal("my-translation-job", props.Get("JobName"))
	a.Equal("IN_PROGRESS", props.Get("JobStatus"))
	a.Equal("en", props.Get("SourceLanguageCode"))
}
