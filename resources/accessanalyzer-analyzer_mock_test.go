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

func Test_Mock_AccessAnalyzer_Filter(t *testing.T) {
	cases := []struct {
		name     string
		analyzer *string
		filtered bool
	}{
		{
			name:     "user analyzer",
			analyzer: ptr.String("example-analyzer"),
			filtered: false,
		},
		{
			name:     "organization analyzer",
			analyzer: ptr.String("ConsoleAnalyzer-ORGANIZATION-1234"),
			filtered: true,
		},
		{
			// Security Hub V2 provisions this analyzer and it cannot be deleted, even after Security Hub V2 is
			// disabled, so it must never be offered up for removal.
			name:     "security hub v2 analyzer",
			analyzer: ptr.String("_AccessAnalyzerForSecurityHubV2-fvw9wfalbdtx"),
			filtered: true,
		},
		{
			// the prefix is only reserved at the start of the name
			name:     "user analyzer that merely mentions security hub v2",
			analyzer: ptr.String("my-AccessAnalyzerForSecurityHubV2"),
			filtered: false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			a := assert.New(t)

			analyzer := &AccessAnalyzer{Name: c.analyzer}

			err := analyzer.Filter()
			if c.filtered {
				a.Error(err)
			} else {
				a.NoError(err)
			}
		})
	}
}
