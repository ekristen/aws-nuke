package resources

import (
	"testing"

	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
)

func Test_Mock_AccessAnalyzer_Properties(t *testing.T) {
	resource := AccessAnalyzer{
		ARN:    ptr.String("arn:aws:accessanalyzer:us-west-2:123456789012:analyzer/1234567890123"),
		Name:   ptr.String("example-analyzer"),
		Status: ptr.String("ACTIVE"),
		Tags: map[string]*string{
			"key": ptr.String("example-key"),
		},
	}

	props := resource.Properties()

	assert.Equal(t, *resource.ARN, props.Get("ARN"))
	assert.Equal(t, "example-analyzer", props.Get("Name"))
	assert.Equal(t, "ACTIVE", props.Get("Status"))
	assert.Equal(t, "example-key", props.Get("tag:key"))
}
