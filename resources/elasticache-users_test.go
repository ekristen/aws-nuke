package resources

import (
	"testing"

	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
)

func Test_ElasticacheUser_Filter(t *testing.T) {
	cases := []struct {
		name     string
		resource *ElasticacheUser
		filtered bool
	}{
		{name: "regular user", resource: &ElasticacheUser{userID: ptr.String("my-custom-user")}, filtered: false},
		{name: "default user", resource: &ElasticacheUser{userID: ptr.String("default")}, filtered: true},
		{name: "default iam user", resource: &ElasticacheUser{userID: ptr.String("default.iam-user")}, filtered: true},
		{name: "id merely starting with default", resource: &ElasticacheUser{userID: ptr.String("default-team")}, filtered: false},
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
