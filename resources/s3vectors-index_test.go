package resources

import (
	"testing"

	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
)

func Test_S3VectorsIndex_Properties(t *testing.T) {
	index := &S3VectorsIndex{
		BucketName: ptr.String("my-vector-bucket"),
		IndexName:  ptr.String("embeddings-index"),
		IndexARN:   ptr.String("arn:aws:s3vectors:us-east-1:123456789012:index/my-vector-bucket/embeddings-index"),
	}

	properties := index.Properties()

	assert.Equal(t, "my-vector-bucket", properties.Get("BucketName"))
	assert.Equal(t, "embeddings-index", properties.Get("IndexName"))
	assert.Equal(t, "arn:aws:s3vectors:us-east-1:123456789012:index/my-vector-bucket/embeddings-index", properties.Get("IndexARN"))
}

func Test_S3VectorsIndex_String(t *testing.T) {
	index := &S3VectorsIndex{
		BucketName: ptr.String("test-bucket"),
		IndexName:  ptr.String("test-index"),
	}

	assert.Equal(t, "test-index", index.String())
}
