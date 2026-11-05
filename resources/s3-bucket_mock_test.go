package resources

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/gotidy/ptr"
	"github.com/sirupsen/logrus"
	logtest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aws/aws-sdk-go-v2/aws"

	libsettings "github.com/ekristen/libnuke/pkg/settings"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const testTagDeniedBucket = "tag-denied-bucket"

// s3BucketTransport answers the S3 REST API in memory for a single empty bucket, whose
// GetBucketTagging fails with tagError ("" for readable tags, "transport" for a network
// failure) until its bucket policy is deleted.
type s3BucketTransport struct {
	tagError   string
	policyGone bool
	bucketGone bool
	calls      []string
}

func s3XMLResponse(request *http.Request, status int, body string) (*http.Response, error) {
	return &http.Response{
		StatusCode: status,
		Header:     http.Header{"Content-Type": {"application/xml"}},
		Body:       io.NopCloser(strings.NewReader(body)),
		Request:    request,
	}, nil
}

func s3APIError(request *http.Request, code string) (*http.Response, error) {
	status := http.StatusForbidden
	switch code {
	case "NoSuchTagSet", "NoSuchBucket", "ObjectLockConfigurationNotFoundError":
		status = http.StatusNotFound
	case "SlowDown":
		status = http.StatusServiceUnavailable
	case "InternalError":
		status = http.StatusInternalServerError
	}
	return s3XMLResponse(request, status, "<Error><Code>"+code+"</Code><Message>test</Message></Error>")
}

// s3BucketOperation names the S3 operation a request makes: its method, and the subresource it
// addresses, if any.
func s3BucketOperation(request *http.Request) string {
	query := request.URL.Query()
	for _, subresource := range []string{"object-lock", "tagging", "policy", "logging", "versions"} {
		if query.Has(subresource) {
			return request.Method + " ?" + subresource
		}
	}
	if query.Get("list-type") == "2" {
		return request.Method + " ?list-type=2"
	}
	return request.Method
}

func (s *s3BucketTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	if strings.HasPrefix(request.URL.Host, "s3.") {
		return s.listBuckets(request)
	}
	if !strings.HasPrefix(request.URL.Host, testTagDeniedBucket+".") {
		return nil, errors.New("unexpected host: " + request.URL.Host)
	}

	switch operation := s3BucketOperation(request); operation {
	case "GET ?object-lock":
		return s3APIError(request, "ObjectLockConfigurationNotFoundError")
	case "GET ?tagging":
		s.calls = append(s.calls, "GetBucketTagging")
		return s.bucketTagging(request)
	case "DELETE ?policy":
		s.calls = append(s.calls, "DeleteBucketPolicy")
		s.policyGone = true
		return s3XMLResponse(request, http.StatusNoContent, "")
	case "PUT ?logging":
		s.calls = append(s.calls, "PutBucketLogging")
		return s3XMLResponse(request, http.StatusOK, "")
	case "GET ?versions":
		s.calls = append(s.calls, "ListObjectVersions")
		return s3XMLResponse(request, http.StatusOK, "<ListVersionsResult><IsTruncated>false</IsTruncated></ListVersionsResult>")
	case "GET ?list-type=2":
		s.calls = append(s.calls, "ListObjectsV2")
		return s3XMLResponse(request, http.StatusOK, "<ListBucketResult><IsTruncated>false</IsTruncated></ListBucketResult>")
	case "DELETE":
		s.calls = append(s.calls, "DeleteBucket")
		s.bucketGone = true
		return s3XMLResponse(request, http.StatusNoContent, "")
	default:
		return nil, errors.New("unexpected request: " + operation + " " + request.URL.String())
	}
}

func (s *s3BucketTransport) listBuckets(request *http.Request) (*http.Response, error) {
	s.calls = append(s.calls, "ListBuckets")
	buckets := ""
	if !s.bucketGone {
		buckets = "<Bucket><Name>" + testTagDeniedBucket + "</Name><BucketRegion>us-west-2</BucketRegion></Bucket>"
	}
	return s3XMLResponse(request, http.StatusOK, "<ListAllMyBucketsResult><Buckets>"+buckets+"</Buckets></ListAllMyBucketsResult>")
}

func (s *s3BucketTransport) bucketTagging(request *http.Request) (*http.Response, error) {
	if s.policyGone || s.tagError == "" {
		return s3XMLResponse(request, http.StatusOK, "<Tagging><TagSet><Tag><Key>owner</Key><Value>team</Value></Tag></TagSet></Tagging>")
	}
	if s.tagError == "transport" {
		return nil, errors.New("connection reset by peer")
	}
	return s3APIError(request, s.tagError)
}

// testS3BucketListerOpts returns lister options that send every request to transport, and the
// hook that collects what the lister logs.
func testS3BucketListerOpts(transport http.RoundTripper) (*nuke.ListerOpts, *logtest.Hook) {
	logger, hook := logtest.NewNullLogger()
	return &nuke.ListerOpts{
		Config: &aws.Config{
			Region:           "us-west-2",
			Credentials:      aws.AnonymousCredentials{},
			HTTPClient:       &http.Client{Transport: transport},
			RetryMaxAttempts: 1,
		},
		Region:    &nuke.Region{Name: "us-west-2"},
		AccountID: ptr.String("012345678901"),
		Logger:    logrus.NewEntry(logger),
	}, hook
}

func Test_Mock_S3Bucket_ListUnreadableTags(t *testing.T) {
	for name, tc := range map[string]struct {
		tagError string
		warning  string // what the logged error says, or "" when nothing is logged
	}{
		"readable tags":   {tagError: ""},
		"no tags":         {tagError: "NoSuchTagSet"},
		"access denied":   {tagError: "AccessDenied", warning: "AccessDenied"},
		"throttling":      {tagError: "SlowDown", warning: "SlowDown"},
		"internal error":  {tagError: "InternalError", warning: "InternalError"},
		"network failure": {tagError: "transport", warning: "connection reset by peer"},
	} {
		t.Run(name, func(t *testing.T) {
			opts, hook := testS3BucketListerOpts(&s3BucketTransport{tagError: tc.tagError})
			resources, err := (&S3BucketLister{}).List(t.Context(), opts)
			require.NoError(t, err)
			require.Len(t, resources, 1, "a failed tag read must not drop the bucket")

			properties := resources[0].(*S3Bucket).Properties()
			assert.Equal(t, testTagDeniedBucket, properties.Get("Name"))
			if tc.tagError == "" {
				assert.Equal(t, "team", properties.Get("tag:owner"))
			} else {
				assert.Empty(t, properties.Get("tag:owner"))
			}

			if tc.warning == "" {
				assert.Empty(t, hook.AllEntries())
				return
			}
			require.Len(t, hook.AllEntries(), 1)
			entry := hook.LastEntry()
			assert.Equal(t, logrus.WarnLevel, entry.Level)
			assert.Contains(t, entry.Message, testTagDeniedBucket)
			logged, ok := entry.Data[logrus.ErrorKey].(error)
			require.True(t, ok)
			assert.ErrorContains(t, logged, tc.warning)
		})
	}
}

func Test_Mock_S3Bucket_ListDeletedBucket(t *testing.T) {
	opts, hook := testS3BucketListerOpts(&s3BucketTransport{tagError: "NoSuchBucket"})
	resources, err := (&S3BucketLister{}).List(t.Context(), opts)
	require.NoError(t, err)
	assert.Empty(t, resources)
	assert.Empty(t, hook.AllEntries())
}

func Test_Mock_S3Bucket_RemoveTagDeniedBucket(t *testing.T) {
	transport := &s3BucketTransport{tagError: "AccessDenied"}
	opts, _ := testS3BucketListerOpts(transport)

	resources, err := (&S3BucketLister{}).List(t.Context(), opts)
	require.NoError(t, err)
	require.Len(t, resources, 1)

	bucket := resources[0].(*S3Bucket)
	bucket.Settings(&libsettings.Setting{})
	require.NoError(t, bucket.Remove(t.Context()))

	assert.Equal(t, []string{
		"ListBuckets", "GetBucketTagging",
		"DeleteBucketPolicy", "PutBucketLogging", "ListObjectVersions", "ListObjectsV2", "DeleteBucket",
	}, transport.calls)

	resources, err = (&S3BucketLister{}).List(t.Context(), opts)
	require.NoError(t, err)
	assert.Empty(t, resources)
}
