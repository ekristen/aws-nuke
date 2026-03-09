package resources

import (
	"testing"
	"time"

	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
)

func Test_ConnectRule_Properties(t *testing.T) {
	a := assert.New(t)

	createdAt := time.Now().Add(-24 * time.Hour)
	updatedAt := time.Now()

	resource := ConnectRule{
		InstanceID:      ptr.String("instance-id"),
		RuleID:          ptr.String("rule-id"),
		Name:            ptr.String("test-rule"),
		PublishStatus:   "PUBLISHED",
		EventSourceName: "OnPostCallAnalysisAvailable",
		CreatedAt:       &createdAt,
		UpdatedAt:       &updatedAt,
	}

	props := resource.Properties()

	a.Equal("instance-id", props.Get("InstanceID"))
	a.Equal("rule-id", props.Get("RuleID"))
	a.Equal("test-rule", props.Get("Name"))
	a.Equal("PUBLISHED", props.Get("PublishStatus"))
	a.Equal("OnPostCallAnalysisAvailable", props.Get("EventSourceName"))
	a.Equal(createdAt.Format(time.RFC3339), props.Get("CreatedAt"))
	a.Equal(updatedAt.Format(time.RFC3339), props.Get("UpdatedAt"))
}

func Test_ConnectRule_String(t *testing.T) {
	a := assert.New(t)

	resource := ConnectRule{
		Name: ptr.String("test-rule"),
	}

	a.Equal("test-rule", resource.String())
}
