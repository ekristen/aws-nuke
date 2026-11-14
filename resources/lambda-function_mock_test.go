package resources

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"

	"github.com/ekristen/libnuke/pkg/filter"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

func newLambdaTestServer(t *testing.T, lastModified string) *httptest.Server {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.HasPrefix(r.URL.Path, "/2015-03-31/functions"):
			_ = json.NewEncoder(w).Encode(map[string]any{
				"Functions": []map[string]any{{
					"FunctionName": "example",
					"FunctionArn":  "arn:aws:lambda:us-east-1:123456789012:function:example",
					"LastModified": lastModified,
				}},
			})
		case strings.HasPrefix(r.URL.Path, "/2017-03-31/tags/"):
			_ = json.NewEncoder(w).Encode(map[string]any{"Tags": map[string]string{"lambda-owner": "platform"}})
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	return server
}

func listLambdaFunctions(t *testing.T, lastModified string) []*LambdaFunction {
	t.Helper()

	server := newLambdaTestServer(t, lastModified)
	cfg := aws.Config{
		Region:       "us-east-1",
		BaseEndpoint: aws.String(server.URL),
		Credentials:  credentials.NewStaticCredentialsProvider("AKID", "SECRET", ""),
	}

	lister := &LambdaFunctionLister{}
	resources, err := lister.List(context.TODO(), &nuke.ListerOpts{Config: &cfg})
	require.NoError(t, err)

	functions := make([]*LambdaFunction, 0, len(resources))
	for _, r := range resources {
		functions = append(functions, r.(*LambdaFunction))
	}

	return functions
}

func TestLambdaFunction_LastModifiedIsTimestamp(t *testing.T) {
	functions := listLambdaFunctions(t, "2019-08-14T22:32:35.183+0000")
	require.Len(t, functions, 1)

	props := functions[0].Properties()
	assert.Equal(t, "example", props.Get("Name"))
	assert.Equal(t, "2019-08-14T22:32:35Z", props.Get("LastModified"))
	assert.Equal(t, "platform", props.Get("tag:lambda-owner"))
}

func TestLambdaFunction_LastModifiedDateOlderThanNowFilter(t *testing.T) {
	functions := listLambdaFunctions(t, "2019-08-14T22:32:35.183+0000")
	require.Len(t, functions, 1)

	f := filter.Filter{
		Property: "LastModified",
		Type:     filter.DateOlderThanNow,
		Value:    "-24h",
	}
	require.NoError(t, f.Validate())

	match, err := f.Match(functions[0].Properties().Get("LastModified"))
	require.NoError(t, err)
	assert.True(t, match)
}

func TestLambdaFunction_LastModifiedUnparsable(t *testing.T) {
	functions := listLambdaFunctions(t, "not-a-date")
	require.Len(t, functions, 1)

	assert.Nil(t, functions[0].LastModified)
	assert.Equal(t, "", functions[0].Properties().Get("LastModified"))
}
