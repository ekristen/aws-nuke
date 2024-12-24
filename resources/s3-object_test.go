package resources

import (
	"fmt"
	"testing"
	"time"

	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
)

func TestS3ObjectProperties(t *testing.T) {
	tests := []struct {
		bucket       string
		key          string
		creationDate time.Time
		versionID    string
		isLatest     bool
	}{
		{
			bucket:       "test-bucket",
			key:          "test-key",
			creationDate: time.Now(),
			versionID:    "null",
			isLatest:     true,
		},
		{
			bucket:       "test-bucket",
			key:          "test-key",
			creationDate: time.Now(),
			versionID:    "test-version-id",
			isLatest:     false,
		},
	}

	for _, test := range tests {
		t.Run(test.bucket, func(t *testing.T) {
			obj := &S3Object{
				Bucket:       ptr.String(test.bucket),
				Key:          ptr.String(test.key),
				VersionID:    ptr.String(test.versionID),
				CreationDate: ptr.Time(test.creationDate),
				IsLatest:     ptr.Bool(test.isLatest),
			}

			got := obj.Properties()
			assert.Equal(t, test.bucket, got.Get("Bucket"))
			assert.Equal(t, test.key, got.Get("Key"))
			assert.Equal(t, test.versionID, got.Get("VersionID"))
			assert.Equal(t, test.creationDate.Format(time.RFC3339), got.Get("CreationDate"))

			if test.isLatest {
				assert.Equal(t, "true", got.Get("IsLatest"))
			} else {
				assert.Equal(t, "false", got.Get("IsLatest"))
			}

			uri := fmt.Sprintf("s3://%s/%s", test.bucket, test.key)
			if test.versionID != "" && test.versionID != "null" && !test.isLatest {
				uri = fmt.Sprintf("%s#%s", uri, test.versionID)
			}
			assert.Equal(t, uri, obj.String())
		})
	}
}
