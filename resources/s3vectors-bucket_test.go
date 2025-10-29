package resources

import (
	"testing"

	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
)

func Test_S3vectorsBucket_Properties(t *testing.T) {
	bucket := &S3vectorsBucket{
		Name: ptr.String("my-vector-bucket"),
		ARN:  ptr.String("arn:aws:s3vectors:us-east-1:123456789012:bucket/my-vector-bucket"),
	}

	properties := bucket.Properties()

	assert.Equal(t, "my-vector-bucket", properties.Get("Name"))
	assert.Equal(t, "arn:aws:s3vectors:us-east-1:123456789012:bucket/my-vector-bucket", properties.Get("ARN"))
}

func Test_S3vectorsBucket_String(t *testing.T) {
	bucket := &S3vectorsBucket{
		Name: ptr.String("test-vector-bucket"),
	}

	assert.Equal(t, "test-vector-bucket", bucket.String())
}
