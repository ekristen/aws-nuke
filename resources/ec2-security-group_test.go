package resources

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
)

func Test_EC2SecurityGroup_Properties(t *testing.T) {
	r := &EC2SecurityGroup{
		group: &ec2.SecurityGroup{
			Tags: []*ec2.Tag{
				{
					Key:   ptr.String("Name"),
					Value: ptr.String("test"),
				},
			},
		},
		ID:      ptr.String("sg-12345678"),
		Name:    ptr.String("test"),
		OwnerID: ptr.String("123456789012"),
	}

	props := r.Properties()

	assert.Equal(t, ptr.ToString(r.Name), props.Get("Name"))
	assert.Equal(t, ptr.ToString(r.OwnerID), props.Get("OwnerID"))
	assert.Equal(t, "test", props.Get("tag:Name"))
}
