package resources

import (
	"testing"

	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
)

func Test_S3vectorsVector_Properties(t *testing.T) {
	vector := &S3vectorsVector{
		VectorBucketName: ptr.String("my-vector-bucket"),
		IndexName:        ptr.String("embeddings-index"),
		Key:              ptr.String("document-123"),
	}

	properties := vector.Properties()

	assert.Equal(t, "my-vector-bucket", properties.Get("VectorBucketName"))
	assert.Equal(t, "embeddings-index", properties.Get("IndexName"))
	assert.Equal(t, "document-123", properties.Get("Key"))
}

func Test_S3vectorsVector_String(t *testing.T) {
	vector := &S3vectorsVector{
		VectorBucketName: ptr.String("test-bucket"),
		IndexName:        ptr.String("test-index"),
		Key:              ptr.String("test-key"),
	}

	assert.Equal(t, "test-key", vector.String())
}
