package resources

import (
	"testing"
	"time"

	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"

	"github.com/aws/aws-sdk-go/aws"           //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/elbv2" //nolint:staticcheck
)

func Test_ELBv2LoadBalancer_Properties(t *testing.T) {
	a := assert.New(t)
	now := time.Now()

	resource := ELBv2LoadBalancer{
		Name:        ptr.String("foobar-name"),
		ARN:         ptr.String("foobar-arn"),
		CreatedTime: ptr.Time(now),
		Tags: []*elbv2.Tag{
			{
				Key:   aws.String("Name"),
				Value: aws.String("foobar-name"),
			},
		},
	}

	props := resource.Properties()

	a.Equal("foobar-name", props.Get("Name"))
	a.Equal("foobar-arn", props.Get("ARN"))
	a.Equal(now.Format(time.RFC3339), props.Get("CreatedTime"))
	a.Equal("foobar-name", props.Get("tag:Name"))
}
