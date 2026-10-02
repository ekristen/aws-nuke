package awsutil_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/smithy-go/middleware"

	"github.com/ekristen/aws-nuke/v3/pkg/awsutil"
)

// applyAPIOptions builds a middleware stack the way the SDK does, so the resulting
// stack reflects what a client created from the config would actually run.
func applyAPIOptions(t *testing.T, cfg *aws.Config) *middleware.Stack {
	t.Helper()

	stack := middleware.NewStack("test", func() interface{} { return nil })
	for _, opt := range cfg.APIOptions {
		assert.NoError(t, opt(stack))
	}

	return stack
}

func TestCrossServiceConfig(t *testing.T) {
	cases := []struct {
		name  string
		guard middleware.InitializeMiddleware
	}{
		{name: "regional", guard: awsutil.SkipRegionalForGlobalService{}},
		{name: "global", guard: awsutil.SkipGlobal{}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			guard := tc.guard

			cfg := &aws.Config{
				Region: "us-east-2",
				APIOptions: []func(*middleware.Stack) error{
					func(stack *middleware.Stack) error {
						return stack.Initialize.Add(guard, middleware.After)
					},
				},
			}

			// Sanity check: the guard is present on the config as handed to a lister.
			_, found := applyAPIOptions(t, cfg).Initialize.Get(guard.ID())
			assert.True(t, found, "expected %s on the original config", guard.ID())

			crossCfg := awsutil.CrossServiceConfig(cfg)

			_, found = applyAPIOptions(t, crossCfg).Initialize.Get(guard.ID())
			assert.False(t, found, "expected %s to be removed", guard.ID())

			// The original must be left intact; it is shared across every resource in
			// the region and the guard still applies to their own clients.
			_, found = applyAPIOptions(t, cfg).Initialize.Get(guard.ID())
			assert.True(t, found, "expected %s to survive on the original config", guard.ID())

			// The region is deliberately untouched so that non-default partitions keep
			// resolving global services to their own endpoints.
			assert.Equal(t, cfg.Region, crossCfg.Region)
		})
	}
}

// TestCrossServiceConfig_NoGuards covers the custom endpoint path, where NewConfig
// attaches neither guard and the helper must be a harmless no-op.
func TestCrossServiceConfig_NoGuards(t *testing.T) {
	cfg := &aws.Config{Region: "us-east-2"}

	crossCfg := awsutil.CrossServiceConfig(cfg)

	assert.Equal(t, "us-east-2", crossCfg.Region)
	assert.Len(t, applyAPIOptions(t, crossCfg).Initialize.List(), 0)
}
