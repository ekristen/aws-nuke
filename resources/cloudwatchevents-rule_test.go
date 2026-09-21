package resources

import (
	"testing"

	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
)

func Test_CloudWatchEventsRule_Filter(t *testing.T) {
	cases := []struct {
		name     string
		resource *CloudWatchEventsRule
		filtered bool
	}{
		{
			name: "unmanaged",
			resource: &CloudWatchEventsRule{
				Name: ptr.String("my-rule"),
			},
			filtered: false,
		},
		{
			name: "managed",
			resource: &CloudWatchEventsRule{
				Name:      ptr.String("SIRGuardDutyRule"),
				ManagedBy: ptr.String("security-ir.amazonaws.com"),
			},
			filtered: true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := c.resource.Filter()
			if c.filtered {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func Test_CloudWatchEventsRule_Properties(t *testing.T) {
	r := &CloudWatchEventsRule{
		Name:         ptr.String("SIRGuardDutyRule"),
		ARN:          ptr.String("arn:aws:events:us-east-1:123456789012:rule/SIRGuardDutyRule"),
		State:        ptr.String("ENABLED"),
		EventBusName: ptr.String("default"),
		ManagedBy:    ptr.String("security-ir.amazonaws.com"),
	}

	props := r.Properties()

	assert.Equal(t, "SIRGuardDutyRule", props.Get("Name"))
	assert.Equal(t, "arn:aws:events:us-east-1:123456789012:rule/SIRGuardDutyRule", props.Get("ARN"))
	assert.Equal(t, "ENABLED", props.Get("State"))
	assert.Equal(t, "default", props.Get("EventBusName"))
	assert.Equal(t, "security-ir.amazonaws.com", props.Get("ManagedBy"))
	assert.Equal(t, "Rule: SIRGuardDutyRule", r.String())
}
