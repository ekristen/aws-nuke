package resources

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aws/aws-sdk-go/aws"             //nolint:staticcheck
	"github.com/aws/aws-sdk-go/aws/credentials" //nolint:staticcheck
	"github.com/aws/aws-sdk-go/aws/session"     //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/filter"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const rdsDescribeDBClustersResponse = `<DescribeDBClustersResponse xmlns="http://rds.amazonaws.com/doc/2014-10-31/">
  <DescribeDBClustersResult>
    <DBClusters>
      <DBCluster>
        <DBClusterIdentifier>example-cluster</DBClusterIdentifier>
        <DBClusterArn>arn:aws:rds:us-east-1:123456789012:cluster:example-cluster</DBClusterArn>
        <DeletionProtection>false</DeletionProtection>
        <ClusterCreateTime>2024-03-01T10:20:30.123Z</ClusterCreateTime>
      </DBCluster>
    </DBClusters>
  </DescribeDBClustersResult>
</DescribeDBClustersResponse>`

const rdsListTagsForResourceResponse = `<ListTagsForResourceResponse xmlns="http://rds.amazonaws.com/doc/2014-10-31/">
  <ListTagsForResourceResult>
    <TagList>
      <Tag><Key>cluster-owner</Key><Value>data</Value></Tag>
    </TagList>
  </ListTagsForResourceResult>
</ListTagsForResourceResponse>`

func TestRDSDBCluster_ClusterCreateTime(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, r.ParseForm())
		w.Header().Set("Content-Type", "text/xml")
		switch r.Form.Get("Action") {
		case "DescribeDBClusters":
			_, _ = w.Write([]byte(rdsDescribeDBClustersResponse))
		case "ListTagsForResource":
			_, _ = w.Write([]byte(rdsListTagsForResourceResponse))
		default:
			http.Error(w, "unexpected action", http.StatusBadRequest)
		}
	}))
	defer server.Close()

	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String("us-east-1"),
		Endpoint:    aws.String(server.URL),
		Credentials: credentials.NewStaticCredentials("AKID", "SECRET", ""),
	})
	require.NoError(t, err)

	lister := &RDSDBClusterLister{}
	resources, err := lister.List(context.TODO(), &nuke.ListerOpts{Session: sess})
	require.NoError(t, err)
	require.Len(t, resources, 1)

	props := resources[0].(*RDSDBCluster).Properties()
	assert.Equal(t, "example-cluster", props.Get("Identifier"))
	assert.Equal(t, "data", props.Get("tag:cluster-owner"))
	assert.Equal(t, "2024-03-01T10:20:30Z", props.Get("ClusterCreateTime"))

	f := filter.Filter{
		Property: "ClusterCreateTime",
		Type:     filter.DateOlderThanNow,
		Value:    "-720h",
	}
	require.NoError(t, f.Validate())

	match, err := f.Match(props.Get("ClusterCreateTime"))
	require.NoError(t, err)
	assert.True(t, match)
}
