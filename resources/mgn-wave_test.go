package resources

import (
	"testing"

	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
)

func Test_MGNWave_Properties_MinimalData(t *testing.T) {
	wave := &MGNWave{
		WaveID:               ptr.String("wave-1234567890abcdef0"),
		Arn:                  ptr.String("arn:aws:mgn:us-east-1:123456789012:wave/wave-1234567890abcdef0"),
		Name:                 ptr.String("TestWave"),
		Description:          ptr.String("Test migration wave"),
		IsArchived:           ptr.Bool(false),
		CreationDateTime:     ptr.String("2024-01-01T12:00:00Z"),
		LastModifiedDateTime: ptr.String("2024-01-01T12:00:00Z"),
		Tags:                 map[string]string{},
	}

	properties := wave.Properties()

	assert.Equal(t, "wave-1234567890abcdef0", properties.Get("WaveID"))
	assert.Equal(t, "arn:aws:mgn:us-east-1:123456789012:wave/wave-1234567890abcdef0", properties.Get("Arn"))
	assert.Equal(t, "TestWave", properties.Get("Name"))
	assert.Equal(t, "Test migration wave", properties.Get("Description"))
	assert.Equal(t, "false", properties.Get("IsArchived"))
	assert.Equal(t, "2024-01-01T12:00:00Z", properties.Get("CreationDateTime"))
	assert.Equal(t, "2024-01-01T12:00:00Z", properties.Get("LastModifiedDateTime"))
}

func Test_MGNWave_Properties_WithTags(t *testing.T) {
	wave := &MGNWave{
		WaveID:               ptr.String("wave-1234567890abcdef0"),
		Arn:                  ptr.String("arn:aws:mgn:us-east-1:123456789012:wave/wave-1234567890abcdef0"),
		Name:                 ptr.String("ProductionWave"),
		Description:          ptr.String("Production environment migration wave"),
		IsArchived:           ptr.Bool(true),
		CreationDateTime:     ptr.String("2024-01-01T12:00:00Z"),
		LastModifiedDateTime: ptr.String("2024-01-01T12:00:00Z"),
		Tags: map[string]string{
			"Name":        "ProdWave",
			"Environment": "production",
			"Phase":       "1",
			"Priority":    "high",
		},
	}

	properties := wave.Properties()

	assert.Equal(t, "wave-1234567890abcdef0", properties.Get("WaveID"))
	assert.Equal(t, "ProductionWave", properties.Get("Name"))
	assert.Equal(t, "true", properties.Get("IsArchived"))
	assert.Equal(t, "ProdWave", properties.Get("tag:Name"))
	assert.Equal(t, "production", properties.Get("tag:Environment"))
	assert.Equal(t, "1", properties.Get("tag:Phase"))
	assert.Equal(t, "high", properties.Get("tag:Priority"))
}

func Test_MGNWave_String(t *testing.T) {
	wave := &MGNWave{
		WaveID: ptr.String("wave-1234567890abcdef0"),
	}

	assert.Equal(t, "wave-1234567890abcdef0", wave.String())
}
