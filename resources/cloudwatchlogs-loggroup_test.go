package resources

import (
	"strconv"
	"testing"
	"time"

	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
)

func TestCloudWatchLogsLogGroupProperties(t *testing.T) {
	now := time.Now().UTC()

	r := &CloudWatchLogsLogGroup{
		Name:            ptr.String("test-log-group"),
		CreatedTime:     ptr.Int64(now.Unix()),
		CreationTime:    ptr.Time(now),
		LastEvent:       ptr.Time(now),
		RetentionInDays: 7,
		Tags: map[string]string{
			"Environment": "production",
		},
	}

	properties := r.Properties()
	assert.Equal(t, properties.Get("Name"), "test-log-group")
	assert.Equal(t, properties.Get("CreatedTime"), strconv.Itoa(int(now.Unix())))
	assert.Equal(t, properties.Get("CreationTime"), now.Format(time.RFC3339))
	assert.Equal(t, properties.Get("LastEvent"), now.Format(time.RFC3339))
	assert.Equal(t, properties.Get("RetentionInDays"), "7")
	assert.Equal(t, properties.Get("tag:Environment"), "production")
}
