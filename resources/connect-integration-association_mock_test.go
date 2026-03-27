package resources

import (
	"testing"

	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
)

func Test_ConnectIntegrationAssociation_Properties(t *testing.T) {
	a := assert.New(t)

	resource := ConnectIntegrationAssociation{
		InstanceID:               ptr.String("instance-id"),
		IntegrationAssociationID: ptr.String("assoc-id"),
		IntegrationType:          "EVENT",
		IntegrationARN:           ptr.String("arn:aws:events:us-east-1:123456789012:event-bus/default"),
		SourceApplicationName:    ptr.String("test-app"),
	}

	props := resource.Properties()

	a.Equal("instance-id", props.Get("InstanceID"))
	a.Equal("assoc-id", props.Get("IntegrationAssociationID"))
	a.Equal("EVENT", props.Get("IntegrationType"))
	a.Equal("arn:aws:events:us-east-1:123456789012:event-bus/default", props.Get("IntegrationARN"))
	a.Equal("test-app", props.Get("SourceApplicationName"))
}

func Test_ConnectIntegrationAssociation_String(t *testing.T) {
	a := assert.New(t)

	resource := ConnectIntegrationAssociation{
		IntegrationAssociationID: ptr.String("assoc-id"),
		SourceApplicationName:    ptr.String("test-app"),
	}

	a.Equal("test-app", resource.String())
}

func Test_ConnectIntegrationAssociation_String_NoName(t *testing.T) {
	a := assert.New(t)

	resource := ConnectIntegrationAssociation{
		IntegrationAssociationID: ptr.String("assoc-id"),
	}

	a.Equal("assoc-id", resource.String())
}
