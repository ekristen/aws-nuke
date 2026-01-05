package resources

import (
	"testing"
	"time"

	textracttypes "github.com/aws/aws-sdk-go-v2/service/textract/types"
	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
)

func TestTextractAdapterVersionProperties(t *testing.T) {
	now := time.Now().UTC()

	resource := &TextractAdapterVersion{
		AdapterID:      ptr.String("adapter-1"),
		AdapterVersion: ptr.String("1"),
		Status:         textracttypes.AdapterVersionStatusActive,
		CreationTime:   ptr.Time(now),
	}

	props := resource.Properties()
	assert.Equal(t, "adapter-1", props.Get("AdapterID"))
	assert.Equal(t, "1", props.Get("AdapterVersion"))

	assert.Equal(t, "adapter-1:1", resource.String())
}

func TestTextractAdapterVersionFilter_CreationInProgress(t *testing.T) {
	resource := &TextractAdapterVersion{
		AdapterID:      ptr.String("adapter-1"),
		AdapterVersion: ptr.String("1"),
		Status:         textracttypes.AdapterVersionStatusCreationInProgress,
	}

	err := resource.Filter()
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "CREATION_IN_PROGRESS")
}

func TestTextractAdapterVersionFilter_CreationError(t *testing.T) {
	resource := &TextractAdapterVersion{
		AdapterID:      ptr.String("adapter-1"),
		AdapterVersion: ptr.String("1"),
		Status:         textracttypes.AdapterVersionStatusCreationError,
	}

	err := resource.Filter()
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "CREATION_ERROR")
}

func TestTextractAdapterVersionFilter_Active(t *testing.T) {
	resource := &TextractAdapterVersion{
		AdapterID:      ptr.String("adapter-1"),
		AdapterVersion: ptr.String("1"),
		Status:         textracttypes.AdapterVersionStatusActive,
	}

	err := resource.Filter()
	assert.Nil(t, err)
}
