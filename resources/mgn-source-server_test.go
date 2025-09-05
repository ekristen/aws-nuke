package resources

import (
	"testing"

	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
)

func Test_MGNSourceServer_Properties_MinimalData(t *testing.T) {
	sourceServer := &MGNSourceServer{
		SourceServerID:  ptr.String("s-1234567890abcdef0"),
		Arn:             ptr.String("arn:aws:mgn:us-east-1:123456789012:source-server/s-1234567890abcdef0"),
		ReplicationType: "AGENT_BASED",
		IsArchived:      ptr.Bool(false),
		Tags:            map[string]string{},
		LifeCycleState:  "READY_FOR_LAUNCH",
		Hostname:        ptr.String("test-server"),
		FQDN:            ptr.String("test-server.example.com"),
	}

	properties := sourceServer.Properties()

	assert.Equal(t, "s-1234567890abcdef0", properties.Get("SourceServerID"))
	assert.Equal(t, "arn:aws:mgn:us-east-1:123456789012:source-server/s-1234567890abcdef0", properties.Get("Arn"))
	assert.Equal(t, "test-server", properties.Get("Hostname"))
	assert.Equal(t, "test-server.example.com", properties.Get("FQDN"))
	assert.Equal(t, "AGENT_BASED", properties.Get("ReplicationType"))
	assert.Equal(t, "READY_FOR_LAUNCH", properties.Get("LifeCycleState"))
	assert.Equal(t, "false", properties.Get("IsArchived"))
}

func Test_MGNSourceServer_Properties_WithTags(t *testing.T) {
	sourceServer := &MGNSourceServer{
		SourceServerID:  ptr.String("s-1234567890abcdef0"),
		Arn:             ptr.String("arn:aws:mgn:us-east-1:123456789012:source-server/s-1234567890abcdef0"),
		ReplicationType: "AGENT_BASED",
		IsArchived:      ptr.Bool(false),
		Tags: map[string]string{
			"Name":        "TestSourceServer",
			"Environment": "test",
			"Team":        "migration",
		},
		LifeCycleState: "READY_FOR_LAUNCH",
		Hostname:       ptr.String("test-server"),
		FQDN:           ptr.String("test-server.example.com"),
	}

	properties := sourceServer.Properties()

	assert.Equal(t, "s-1234567890abcdef0", properties.Get("SourceServerID"))
	assert.Equal(t, "TestSourceServer", properties.Get("tag:Name"))
	assert.Equal(t, "test", properties.Get("tag:Environment"))
	assert.Equal(t, "migration", properties.Get("tag:Team"))
}

func Test_MGNSourceServer_String(t *testing.T) {
	sourceServer := &MGNSourceServer{
		SourceServerID: ptr.String("s-1234567890abcdef0"),
	}

	assert.Equal(t, "s-1234567890abcdef0", sourceServer.String())
}
