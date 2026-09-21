package resources

import (
	"testing"

	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
)

func Test_CloudWatchEventsTarget_Filter(t *testing.T) {
	cases := []struct {
		name     string
		resource *CloudWatchEventsTarget
		filtered bool
	}{
		{
			name: "unmanaged",
			resource: &CloudWatchEventsTarget{
				Name:     ptr.String("my-rule"),
				TargetID: ptr.String("my-target"),
			},
			filtered: false,
		},
		{
			name: "managed",
			resource: &CloudWatchEventsTarget{
				Name:      ptr.String("SIRGuardDutyRule"),
				TargetID:  ptr.String("security-ir"),
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

func Test_CloudWatchEventsTarget_Properties(t *testing.T) {
	r := &CloudWatchEventsTarget{
		Name:      ptr.String("SIRGuardDutyRule"),
		TargetID:  ptr.String("security-ir"),
		BusName:   ptr.String("default"),
		ManagedBy: ptr.String("security-ir.amazonaws.com"),
	}

	props := r.Properties()

	assert.Equal(t, "SIRGuardDutyRule", props.Get("Name"))
	assert.Equal(t, "security-ir", props.Get("TargetID"))
	assert.Equal(t, "default", props.Get("BusName"))
	assert.Equal(t, "security-ir.amazonaws.com", props.Get("ManagedBy"))
	assert.Equal(t, "Rule: SIRGuardDutyRule Target ID: security-ir", r.String())
}
