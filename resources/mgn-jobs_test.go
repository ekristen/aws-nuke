package resources

import (
	"testing"

	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
)

func Test_MGNJob_Properties_MinimalData(t *testing.T) {
	job := &MGNJob{
		JobID:            ptr.String("mjb-1234567890abcdef0"),
		Arn:              ptr.String("arn:aws:mgn:us-east-1:123456789012:job/mjb-1234567890abcdef0"),
		Type:             "LAUNCH",
		Status:           "COMPLETED",
		InitiatedBy:      "USER",
		CreationDateTime: ptr.String("2024-01-01T12:00:00Z"),
		Tags:             map[string]string{},
	}

	properties := job.Properties()

	assert.Equal(t, "mjb-1234567890abcdef0", properties.Get("JobID"))
	assert.Equal(t, "arn:aws:mgn:us-east-1:123456789012:job/mjb-1234567890abcdef0", properties.Get("Arn"))
	assert.Equal(t, "LAUNCH", properties.Get("Type"))
	assert.Equal(t, "COMPLETED", properties.Get("Status"))
	assert.Equal(t, "USER", properties.Get("InitiatedBy"))
	assert.Equal(t, "2024-01-01T12:00:00Z", properties.Get("CreationDateTime"))
}

func Test_MGNJob_Properties_WithEndTime(t *testing.T) {
	job := &MGNJob{
		JobID:            ptr.String("mjb-1234567890abcdef0"),
		Arn:              ptr.String("arn:aws:mgn:us-east-1:123456789012:job/mjb-1234567890abcdef0"),
		Type:             "TERMINATE",
		Status:           "COMPLETED",
		InitiatedBy:      "USER",
		CreationDateTime: ptr.String("2024-01-01T12:00:00Z"),
		EndDateTime:      ptr.String("2024-01-01T13:00:00Z"),
		Tags: map[string]string{
			"Name":        "TestJob",
			"Environment": "test",
			"Type":        "migration",
		},
	}

	properties := job.Properties()

	assert.Equal(t, "mjb-1234567890abcdef0", properties.Get("JobID"))
	assert.Equal(t, "TERMINATE", properties.Get("Type"))
	assert.Equal(t, "2024-01-01T13:00:00Z", properties.Get("EndDateTime"))
	assert.Equal(t, "TestJob", properties.Get("tag:Name"))
	assert.Equal(t, "test", properties.Get("tag:Environment"))
	assert.Equal(t, "migration", properties.Get("tag:Type"))
}

func Test_MGNJob_String(t *testing.T) {
	job := &MGNJob{
		JobID: ptr.String("mjb-1234567890abcdef0"),
	}

	assert.Equal(t, "mjb-1234567890abcdef0", job.String())
}
