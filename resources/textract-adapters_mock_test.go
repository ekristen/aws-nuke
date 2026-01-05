package resources

import (
	"testing"
	"time"

	textracttypes "github.com/aws/aws-sdk-go-v2/service/textract/types"
	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
)

func TestTextractAdapterProperties(t *testing.T) {
	now := time.Now().UTC()

	resource := &TextractAdapter{
		AdapterID:    ptr.String("adapter-1"),
		AdapterName:  ptr.String("test-adapter"),
		AutoUpdate:   textracttypes.AutoUpdateEnabled,
		Description:  ptr.String("Test description"),
		CreationTime: ptr.Time(now),
		Tags: map[string]string{
			"Environment": "test",
		},
	}

	props := resource.Properties()
	assert.Equal(t, "adapter-1", props.Get("AdapterID"))
	assert.Equal(t, "test-adapter", props.Get("AdapterName"))
	assert.Equal(t, "test", props.Get("tag:Environment"))

	assert.Equal(t, "adapter-1", resource.String())
}
