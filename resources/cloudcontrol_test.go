package resources

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestCloudControlParseProperties(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)

	cases := []struct {
		name    string
		payload string
		want    []string
	}{
		{
			name:    "ActualEC2VPC",
			payload: `{"VpcId":"vpc-456","InstanceTenancy":"default","CidrBlockAssociations":["vpc-cidr-assoc-1234", "vpc-cidr-assoc-5678"],"CidrBlock":"10.10.0.0/16","Tags":[{"Value":"Kubernetes VPC","Key":"Name"}]}`,
			want: []string{
				`CidrBlock: "10.10.0.0/16"`,
				`Tags.["Name"]: "Kubernetes VPC"`,
				`VpcId: "vpc-456"`,
				`InstanceTenancy: "default"`,
				`CidrBlockAssociations.["vpc-cidr-assoc-1234"]: "true"`,
				`CidrBlockAssociations.["vpc-cidr-assoc-5678"]: "true"`,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := cloudControlParseProperties(tc.payload)
			assert.NoError(t, err)
			for _, w := range tc.want {
				assert.Contains(t, result.String(), w)
			}
		})
	}
}
