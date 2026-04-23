package resources

import (
	"testing"
	"time"

	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
)

func Test_Mock_S3TablesNamespace_Properties(t *testing.T) {
	createdTime := time.Now().Add(-24 * time.Hour)

	r := &S3TablesNamespace{
		Name:            ptr.String("name"),
		CreationDate:    &createdTime,
		TableBucketName: ptr.String("tableBucketName"),
		tableBucketARN:  ptr.String("tableBucketARN"),
	}

	properties := r.Properties()
	assert.Equal(t, "name", properties.Get("Name"))
	assert.Equal(t, createdTime.Format(time.RFC3339), properties.Get("CreationDate"))
	assert.Equal(t, "tableBucketName", properties.Get("TableBucketName"))
	assert.Equal(t, "", properties.Get("tableBucketARN"))
}
