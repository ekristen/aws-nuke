package resources

import (
	"testing"
	"time"

	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
)

func Test_Mock_S3TablesBucket_Properties(t *testing.T) {
	createdTime := time.Now().Add(-24 * time.Hour)

	r := &S3TablesBucket{
		Name:         ptr.String("name"),
		CreationDate: &createdTime,
		Type:         "aws",
	}

	properties := r.Properties()
	assert.Equal(t, "name", properties.Get("Name"))
	assert.Equal(t, createdTime.Format(time.RFC3339), properties.Get("CreationDate"))
	assert.Equal(t, "aws", properties.Get("Type"))
}
