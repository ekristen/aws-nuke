package resources

import (
	"testing"
	"time"

	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
)

func Test_Mock_S3TablesTable_Properties(t *testing.T) {
	createdTime := time.Now().Add(-24 * time.Hour)

	r := &S3TablesTable{
		Name:             ptr.String("name"),
		Namespace:        ptr.String("namespace"),
		CreationDate:     &createdTime,
		TableBucketName:  ptr.String("tableBucketName"),
		tableBucketARN:   ptr.String("tableBucketARN"),
		ManagedByService: ptr.String("metadata.s3.amazonaws.com"),
		Type:             "aws",
	}

	properties := r.Properties()
	assert.Equal(t, "name", properties.Get("Name"))
	assert.Equal(t, "namespace", properties.Get("Namespace"))
	assert.Equal(t, createdTime.Format(time.RFC3339), properties.Get("CreationDate"))
	assert.Equal(t, "tableBucketName", properties.Get("TableBucketName"))
	assert.Equal(t, "", properties.Get("tableBucketARN"))
	assert.Equal(t, "metadata.s3.amazonaws.com", properties.Get("ManagedByService"))
	assert.Equal(t, "aws", properties.Get("Type"))
}
