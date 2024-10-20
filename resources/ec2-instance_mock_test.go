package resources

import (
	"testing"
	"time"

	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"

	"github.com/aws/aws-sdk-go/service/ec2"
)

func Test_Mock_EC2Instance_Properties(t *testing.T) {
	now := time.Now().UTC()
	r := &EC2Instance{
		ID:           ptr.String("instanceID"),
		ImageID:      ptr.String("imageID"),
		State:        ptr.String("state"),
		InstanceType: ptr.String("instanceType"),
		LaunchTime:   ptr.Time(now),
		Tags: []*ec2.Tag{
			{
				Key:   ptr.String("key"),
				Value: ptr.String("value"),
			},
		},
	}

	properties := r.Properties()

	assert.Equal(t, "instanceID", properties.Get("Identifier"))
	assert.Equal(t, "imageID", properties.Get("ImageIdentifier"))
	assert.Equal(t, "state", properties.Get("InstanceState"))
	assert.Equal(t, "instanceType", properties.Get("InstanceType"))
	assert.Equal(t, now.Format(time.RFC3339), properties.Get("LaunchTime"))
	assert.Equal(t, "value", properties.Get("tag:key"))
}
