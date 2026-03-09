package resources

import (
	"testing"

	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"

	connecttypes "github.com/aws/aws-sdk-go-v2/service/connect/types"
)

func Test_ConnectQueue_Properties(t *testing.T) {
	a := assert.New(t)

	resource := ConnectQueue{
		InstanceID: ptr.String("instance-id"),
		QueueID:    ptr.String("queue-id"),
		Name:       ptr.String("test-queue"),
		QueueType:  "STANDARD",
		ARN:        ptr.String("arn:aws:connect:us-east-1:123456789012:instance/instance-id/queue/queue-id"),
		Tags: map[string]string{
			"Environment": "test",
		},
	}

	props := resource.Properties()

	a.Equal("instance-id", props.Get("InstanceID"))
	a.Equal("queue-id", props.Get("QueueID"))
	a.Equal("test-queue", props.Get("Name"))
	a.Equal("STANDARD", props.Get("QueueType"))
	a.Equal("arn:aws:connect:us-east-1:123456789012:instance/instance-id/queue/queue-id", props.Get("ARN"))
	a.Equal("test", props.Get("tag:Environment"))
}

func Test_ConnectQueue_String(t *testing.T) {
	a := assert.New(t)

	resource := ConnectQueue{
		Name: ptr.String("test-queue"),
	}

	a.Equal("test-queue", resource.String())
}

func Test_ConnectQueue_Filter_AgentQueue(t *testing.T) {
	a := assert.New(t)

	resource := ConnectQueue{
		Name:      ptr.String("agent-queue"),
		QueueType: string(connecttypes.QueueTypeAgent),
	}

	err := resource.Filter()
	a.NotNil(err)
	a.Contains(err.Error(), "cannot delete agent queue")
}

func Test_ConnectQueue_Filter_StandardQueue(t *testing.T) {
	a := assert.New(t)

	resource := ConnectQueue{
		Name:      ptr.String("standard-queue"),
		QueueType: string(connecttypes.QueueTypeStandard),
	}

	err := resource.Filter()
	a.Nil(err)
}
