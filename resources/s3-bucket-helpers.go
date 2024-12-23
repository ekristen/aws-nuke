package resources

import (
	"context"
	"time"

	"github.com/gotidy/ptr"
	"github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"

	"github.com/ekristen/aws-nuke/v3/pkg/awsmod"
)

func bypassGovernanceRetention(input *s3.DeleteObjectsInput) {
	input.BypassGovernanceRetention = ptr.Bool(true)
}

type s3DeleteVersionListIterator struct {
	Bucket                    *string
	Paginator                 *s3.ListObjectVersionsPaginator
	objects                   []s3types.ObjectVersion
	lastNotify                time.Time
	BypassGovernanceRetention *bool
	err                       error
}

func newS3DeleteVersionListIterator(
	svc *s3.Client,
	input *s3.ListObjectVersionsInput,
	bypass bool,
	opts ...func(*s3DeleteVersionListIterator)) awsmod.BatchDeleteIterator {
	iter := &s3DeleteVersionListIterator{
		Bucket:                    input.Bucket,
		Paginator:                 s3.NewListObjectVersionsPaginator(svc, input),
		BypassGovernanceRetention: ptr.Bool(bypass),
	}

	for _, opt := range opts {
		opt(iter)
	}

	return iter
}

// Next will use the S3API client to iterate through a list of objects.
func (iter *s3DeleteVersionListIterator) Next() bool {
	if len(iter.objects) > 0 {
		iter.objects = iter.objects[1:]
		if len(iter.objects) > 0 {
			return true
		}
	}

	if !iter.Paginator.HasMorePages() {
		return false
	}

	page, err := iter.Paginator.NextPage(context.TODO())
	if err != nil {
		iter.err = err
		return false
	}

	iter.objects = page.Versions
	for _, entry := range page.DeleteMarkers {
		iter.objects = append(iter.objects, s3types.ObjectVersion{
			Key:       entry.Key,
			VersionId: entry.VersionId,
		})
	}

	if len(iter.objects) > 500 && (iter.lastNotify.IsZero() || time.Since(iter.lastNotify) > 120*time.Second) {
		logrus.Infof(
			"S3Bucket: %s - empty bucket operation in progress, this could take a while, please be patient",
			*iter.Bucket)
		iter.lastNotify = time.Now().UTC()
	}

	return len(iter.objects) > 0
}

// Err will return the last known error from Next.
func (iter *s3DeleteVersionListIterator) Err() error {
	return iter.err
}

// DeleteObject will return the current object to be deleted.
func (iter *s3DeleteVersionListIterator) DeleteObject() awsmod.BatchDeleteObject {
	return awsmod.BatchDeleteObject{
		Object: &s3.DeleteObjectInput{
			Bucket:                    iter.Bucket,
			Key:                       iter.objects[0].Key,
			VersionId:                 iter.objects[0].VersionId,
			BypassGovernanceRetention: iter.BypassGovernanceRetention,
		},
	}
}

type s3ObjectDeleteListIterator struct {
	Bucket                    *string
	Paginator                 *s3.ListObjectsV2Paginator
	objects                   []s3types.Object
	lastNotify                time.Time
	BypassGovernanceRetention bool
	err                       error
}

func newS3ObjectDeleteListIterator(
	svc *s3.Client,
	input *s3.ListObjectsV2Input,
	bypass bool,
	opts ...func(*s3ObjectDeleteListIterator)) awsmod.BatchDeleteIterator {
	iter := &s3ObjectDeleteListIterator{
		Bucket:                    input.Bucket,
		Paginator:                 s3.NewListObjectsV2Paginator(svc, input),
		BypassGovernanceRetention: bypass,
	}

	for _, opt := range opts {
		opt(iter)
	}
	return iter
}

// Next will use the S3API client to iterate through a list of objects.
func (iter *s3ObjectDeleteListIterator) Next() bool {
	if len(iter.objects) > 0 {
		iter.objects = iter.objects[1:]
		if len(iter.objects) > 0 {
			return true
		}
	}

	if !iter.Paginator.HasMorePages() {
		return false
	}

	page, err := iter.Paginator.NextPage(context.TODO())
	if err != nil {
		iter.err = err
		return false
	}

	iter.objects = page.Contents

	if len(iter.objects) > 500 && (iter.lastNotify.IsZero() || time.Since(iter.lastNotify) > 120*time.Second) {
		logrus.Infof(
			"S3Bucket: %s - empty bucket operation in progress, this could take a while, please be patient",
			*iter.Bucket)
		iter.lastNotify = time.Now().UTC()
	}

	return len(iter.objects) > 0
}

// Err will return the last known error from Next.
func (iter *s3ObjectDeleteListIterator) Err() error {
	return iter.err
}

// DeleteObject will return the current object to be deleted.
func (iter *s3ObjectDeleteListIterator) DeleteObject() awsmod.BatchDeleteObject {
	return awsmod.BatchDeleteObject{
		Object: &s3.DeleteObjectInput{
			Bucket:                    iter.Bucket,
			Key:                       iter.objects[0].Key,
			BypassGovernanceRetention: ptr.Bool(iter.BypassGovernanceRetention),
		},
	}
}
