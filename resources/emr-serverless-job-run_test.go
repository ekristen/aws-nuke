package resources

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/emrserverless/types"
	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
)

func Test_EMRServerlessJobRun_Properties(t *testing.T) {
	now := time.Now()

	jobRun := &EMRServerlessJobRun{
		ApplicationID:   ptr.String("app-123456"),
		ApplicationName: ptr.String("my-spark-app"),
		JobRunID:        ptr.String("jr-987654321"),
		Name:            ptr.String("daily-etl-job"),
		ARN:             ptr.String("arn:aws:emr-serverless:us-east-1:123456789012:application/app-123456/jobruns/jr-987654321"),
		State:           types.JobRunStateRunning,
		CreatedAt:       &now,
		UpdatedAt:       &now,
		Tags: map[string]string{
			"Environment": "production",
			"Team":        "data-platform",
		},
	}

	properties := jobRun.Properties()

	assert.Equal(t, "app-123456", properties.Get("ApplicationID"))
	assert.Equal(t, "my-spark-app", properties.Get("ApplicationName"))
	assert.Equal(t, "jr-987654321", properties.Get("JobRunID"))
	assert.Equal(t, "daily-etl-job", properties.Get("Name"))
	assert.Equal(t, "arn:aws:emr-serverless:us-east-1:123456789012:application/app-123456/jobruns/jr-987654321", properties.Get("ARN"))
	assert.Equal(t, "RUNNING", properties.Get("State"))
	assert.Equal(t, now.Format(time.RFC3339), properties.Get("CreatedAt"))
	assert.Equal(t, now.Format(time.RFC3339), properties.Get("UpdatedAt"))
	assert.Equal(t, "production", properties.Get("tag:Environment"))
	assert.Equal(t, "data-platform", properties.Get("tag:Team"))
}

func Test_EMRServerlessJobRun_String(t *testing.T) {
	jobRun := &EMRServerlessJobRun{
		JobRunID: ptr.String("jr-abcd1234"),
		Name:     ptr.String("test-job"),
	}

	assert.Equal(t, "jr-abcd1234", jobRun.String())
}

func Test_EMRServerlessJobRun_Filter_NotCancellable(t *testing.T) {
	jobRun := &EMRServerlessJobRun{
		JobRunID: ptr.String("jr-123"),
		State:    types.JobRunStateSuccess,
	}

	err := jobRun.Filter()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not in a cancellable state")
}

func Test_EMRServerlessJobRun_Filter_Cancellable(t *testing.T) {
	testCases := []types.JobRunState{
		types.JobRunStateSubmitted,
		types.JobRunStatePending,
		types.JobRunStateScheduled,
		types.JobRunStateRunning,
	}

	for _, state := range testCases {
		jobRun := &EMRServerlessJobRun{
			JobRunID: ptr.String("jr-123"),
			State:    state,
		}

		err := jobRun.Filter()
		assert.NoError(t, err, "Expected no error for state: %s", state)
	}
}

func Test_isJobRunCancellable(t *testing.T) {
	testCases := []struct {
		state       types.JobRunState
		cancellable bool
	}{
		{types.JobRunStateSubmitted, true},
		{types.JobRunStatePending, true},
		{types.JobRunStateScheduled, true},
		{types.JobRunStateRunning, true},
		{types.JobRunStateSuccess, false},
		{types.JobRunStateFailed, false},
		{types.JobRunStateCancelling, false},
		{types.JobRunStateCancelled, false},
	}

	for _, tc := range testCases {
		result := isJobRunCancellable(tc.state)
		assert.Equal(t, tc.cancellable, result, "State: %s", tc.state)
	}
}
