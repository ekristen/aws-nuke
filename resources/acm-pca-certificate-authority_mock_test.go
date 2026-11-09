package resources

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aws/aws-sdk-go/aws"             //nolint:staticcheck
	"github.com/aws/aws-sdk-go/aws/awserr"      //nolint:staticcheck
	"github.com/aws/aws-sdk-go/aws/credentials" //nolint:staticcheck
	"github.com/aws/aws-sdk-go/aws/session"     //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/acmpca"  //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

func Test_Mock_ACMPCACertificateAuthority_Properties(t *testing.T) {
	r := &ACMPCACertificateAuthority{
		ARN:    ptr.String("certificateAuthorityARN"),
		Status: ptr.String("status"),
		Tags: []*acmpca.Tag{
			{
				Key:   ptr.String("key"),
				Value: ptr.String("value"),
			},
		},
	}

	properties := r.Properties()
	assert.Equal(t, "value", properties.Get("tag:key"))
	assert.Equal(t, "status", properties.Get("Status"))
	assert.Equal(t, "certificateAuthorityARN", properties.Get("ARN"))
	assert.Equal(t, "certificateAuthorityARN", r.String())
}

const testDeletedCAMessage = "The certificate authority is in the DELETED state and must be restored to complete this action."

// acmpcaTransport answers the ACM PCA JSON API in memory: one CA per ARN, with
// the error, if any, that ListTags returns for it.
type acmpcaTransport struct {
	authorities []map[string]string
	tagErrors   map[string]string
}

func (t *acmpcaTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	var input map[string]any
	if err := json.NewDecoder(request.Body).Decode(&input); err != nil {
		return nil, err
	}
	status, body := http.StatusOK, map[string]any{}
	switch request.Header.Get("X-Amz-Target") {
	case "ACMPrivateCA.ListCertificateAuthorities":
		body["CertificateAuthorities"] = t.authorities
	case "ACMPrivateCA.ListTags":
		arn, _ := input["CertificateAuthorityArn"].(string)
		if code, ok := t.tagErrors[arn]; ok {
			status, body = http.StatusBadRequest, map[string]any{"__type": code, "message": testDeletedCAMessage}
		} else {
			body["Tags"] = []map[string]string{{"Key": "owner", "Value": "team"}}
		}
	default:
		return nil, assert.AnError
	}
	content, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	return &http.Response{
		StatusCode: status,
		Header:     http.Header{"Content-Type": {"application/x-amz-json-1.1"}},
		Body:       io.NopCloser(bytes.NewReader(content)),
		Request:    request,
	}, nil
}

func testACMPCAListerOpts(t *testing.T, transport http.RoundTripper) *nuke.ListerOpts {
	t.Helper()
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String("us-east-2"),
		Credentials: credentials.AnonymousCredentials,
		MaxRetries:  aws.Int(0),
		HTTPClient:  &http.Client{Transport: transport},
	})
	require.NoError(t, err)
	return &nuke.ListerOpts{Session: sess}
}

func Test_Mock_ACMPCACertificateAuthority_ListDeleted(t *testing.T) {
	active := "arn:aws:acm-pca:us-east-2:012345678901:certificate-authority/active"
	deleted := "arn:aws:acm-pca:us-east-2:012345678901:certificate-authority/deleted"

	for name, lister := range map[string]registry.Lister{
		ACMPCACertificateAuthorityResource:      &ACMPCACertificateAuthorityLister{},
		ACMPCACertificateAuthorityStateResource: &ACMPCACertificateAuthorityStateLister{},
	} {
		t.Run(name, func(t *testing.T) {
			transport := &acmpcaTransport{
				authorities: []map[string]string{
					{"Arn": deleted, "Status": acmpca.CertificateAuthorityStatusDeleted},
					{"Arn": active, "Status": acmpca.CertificateAuthorityStatusActive},
				},
				tagErrors: map[string]string{deleted: acmpca.ErrCodeInvalidStateException},
			}

			resources, err := lister.List(t.Context(), testACMPCAListerOpts(t, transport))
			require.NoError(t, err)
			require.Len(t, resources, 2)

			assert.EqualError(t, resources[0].(resource.Filter).Filter(), "already deleted")
			assert.Equal(t, "", resources[0].(resource.PropertyGetter).Properties().Get("tag:owner"))
			assert.Equal(t, "team", resources[1].(resource.PropertyGetter).Properties().Get("tag:owner"))
		})
	}
}

func Test_Mock_ACMPCACertificateAuthority_ListTagErrors(t *testing.T) {
	arn := "arn:aws:acm-pca:us-east-2:012345678901:certificate-authority/ca"

	for name, tc := range map[string]struct {
		status string
		code   string
	}{
		"invalid state of an active CA":  {acmpca.CertificateAuthorityStatusActive, acmpca.ErrCodeInvalidStateException},
		"access denied for a deleted CA": {acmpca.CertificateAuthorityStatusDeleted, "AccessDeniedException"},
	} {
		for lister := range map[registry.Lister]bool{
			&ACMPCACertificateAuthorityLister{}:      true,
			&ACMPCACertificateAuthorityStateLister{}: true,
		} {
			t.Run(name, func(t *testing.T) {
				transport := &acmpcaTransport{
					authorities: []map[string]string{{"Arn": arn, "Status": tc.status}},
					tagErrors:   map[string]string{arn: tc.code},
				}

				resources, err := lister.List(t.Context(), testACMPCAListerOpts(t, transport))
				assert.Nil(t, resources)
				var awsErr awserr.Error
				require.ErrorAs(t, err, &awsErr)
				assert.Equal(t, tc.code, awsErr.Code())
			})
		}
	}
}
