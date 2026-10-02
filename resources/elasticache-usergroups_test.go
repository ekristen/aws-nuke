package resources

import (
	"testing"

	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
)

func Test_ElasticacheUserGroup_Filter(t *testing.T) {
	cases := []struct {
		name     string
		resource *ElasticacheUserGroup
		filtered bool
	}{
		{name: "regular group", resource: &ElasticacheUserGroup{groupID: ptr.String("my-custom-group")}, filtered: false},
		{name: "default iam user group", resource: &ElasticacheUserGroup{groupID: ptr.String("default.iam-user-group")}, filtered: true},
		{name: "id merely starting with default", resource: &ElasticacheUserGroup{groupID: ptr.String("default-team-group")}, filtered: false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := c.resource.Filter()
			if c.filtered {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
