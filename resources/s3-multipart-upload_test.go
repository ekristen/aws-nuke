package resources

import (
	"fmt"
	"testing"

	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
)

func TestS3MultipartUploadProperties(t *testing.T) {
	tests := []struct {
		bucket   string
		key      string
		uploadID string
	}{
		{
			bucket:   "test-bucket",
			key:      "test-key",
			uploadID: "test-upload-id",
		},
	}

	for _, test := range tests {
		t.Run(test.bucket, func(t *testing.T) {
			obj := &S3MultipartUpload{
				Bucket:   ptr.String(test.bucket),
				Key:      ptr.String(test.key),
				UploadID: ptr.String(test.uploadID),
			}

			got := obj.Properties()
			assert.Equal(t, test.bucket, got.Get("Bucket"))
			assert.Equal(t, test.key, got.Get("Key"))
			assert.Equal(t, test.uploadID, got.Get("UploadID"))

			uri := fmt.Sprintf("s3://%s/%s#%s", test.bucket, test.key, test.uploadID)
			assert.Equal(t, uri, obj.String())
		})
	}
}
