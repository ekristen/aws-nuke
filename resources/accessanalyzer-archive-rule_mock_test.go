package resources

import (
	"testing"

	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
)

func Test_Mock_AccessAnalyzerArchiveRule_Properties(t *testing.T) {
	resource := ArchiveRule{
		RuleName:     ptr.String("example-rule"),
		AnalyzerName: ptr.String("example-archive"),
	}

	props := resource.Properties()
	assert.Equal(t, "example-rule", props.Get("RuleName"))
	assert.Equal(t, "example-archive", props.Get("AnalyzerName"))
}
