package awsutil

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httputil"
	"regexp"

	"github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws/awserr"

	"github.com/ekristen/libnuke/pkg/utils"
)

var (
	RESecretHeader = regexp.MustCompile(`(?m:^([^:]*(Auth|Security)[^:]*):.*$)`)
)

func HideSecureHeaders(dump []byte) []byte {
	return RESecretHeader.ReplaceAll(dump, []byte("$1: <hidden>"))
}

func DumpRequest(r *http.Request) string {
	dump, err := httputil.DumpRequest(r, true)
	if err != nil {
		logrus.WithField("Error", err).Warnf("failed to dump HTTP request")
		return ""
	}

	dump = bytes.TrimSpace(dump)
	dump = HideSecureHeaders(dump)
	dump = utils.IndentBytes(dump, []byte("    > "))
	return string(dump)
}

func DumpResponse(r *http.Response) string {
	dump, err := httputil.DumpResponse(r, true)
	if err != nil {
		logrus.WithField("Error", err).Warnf("failed to dump HTTP response")
		return ""
	}

	dump = bytes.TrimSpace(dump)
	dump = utils.IndentBytes(dump, []byte("    < "))
	return string(dump)
}

func IsAWSError(err error, code string) bool {
	var aerr awserr.Error
	ok := errors.As(err, &aerr)
	if !ok {
		return false
	}

	return aerr.Code() == code
}
