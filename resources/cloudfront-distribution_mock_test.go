package resources

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront/types"

	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

func Test_CloudFrontDistribution_List(t *testing.T) {
	mockSvc := new(mockCloudFrontClient)
	mockSvc.On("ListDistributions", mock.Anything, mock.Anything).Return(&cloudfront.ListDistributionsOutput{
		DistributionList: &types.DistributionList{
			Items: []types.DistributionSummary{
				{Id: ptr.String("test-id"), ARN: ptr.String("test-arn"), Status: ptr.String("Deployed"), LastModifiedTime: ptr.Time(time.Now())},
			},
			IsTruncated: ptr.Bool(false),
		},
	}, nil)
	mockSvc.On("ListTagsForResource", mock.Anything, mock.Anything).Return(&cloudfront.ListTagsForResourceOutput{
		Tags: &types.Tags{
			Items: []types.Tag{
				{Key: ptr.String("test-key"), Value: ptr.String("test-value")},
			},
		},
	}, nil)

	lister := &CloudFrontDistributionLister{
		mockSvc: mockSvc,
	}
	opts := &nuke.ListerOpts{Config: &aws.Config{
		Region: "us-east-1",
	}}
	resources, err := lister.List(context.TODO(), opts)
	assert.NoError(t, err)
	assert.Len(t, resources, 1)
}

func Test_CloudFrontDistribution_Remove(t *testing.T) {
	mockSvc := new(mockCloudFrontClient)
	mockSvc.On("GetDistributionConfig", mock.Anything, mock.Anything).Return(&cloudfront.GetDistributionConfigOutput{
		DistributionConfig: &types.DistributionConfig{Enabled: ptr.Bool(true)},
		ETag:               ptr.String("test-etag"),
	}, nil)
	mockSvc.On("UpdateDistribution", mock.Anything, mock.Anything).Return(&cloudfront.UpdateDistributionOutput{}, nil)
	mockSvc.On("DeleteDistribution", mock.Anything, mock.Anything).Return(&cloudfront.DeleteDistributionOutput{}, nil)

	r := &CloudFrontDistribution{
		svc: mockSvc,
		ID:  ptr.String("test-id"),
	}
	err := r.Remove(context.TODO())
	assert.NoError(t, err)
}

func Test_CloudFrontDistribution_RemoveGone(t *testing.T) {
	mockSvc := new(mockCloudFrontClient)
	mockSvc.On("GetDistributionConfig", mock.Anything, mock.Anything).Return(&cloudfront.GetDistributionConfigOutput{
		DistributionConfig: &types.DistributionConfig{Enabled: ptr.Bool(true)},
		ETag:               ptr.String("test-etag"),
	}, nil)
	mockSvc.On("UpdateDistribution", mock.Anything, mock.Anything).Return(&cloudfront.UpdateDistributionOutput{}, nil)
	mockSvc.On("DeleteDistribution", mock.Anything, mock.Anything).Return(&cloudfront.DeleteDistributionOutput{}, nil)

	mockSvc.On("ListDistributions", mock.Anything, mock.Anything).Return(&cloudfront.ListDistributionsOutput{
		DistributionList: &types.DistributionList{
			Items:       []types.DistributionSummary{},
			IsTruncated: ptr.Bool(false),
		},
	}, nil)

	r := &CloudFrontDistribution{
		svc: mockSvc,
		ID:  ptr.String("test-id"),
	}
	err := r.Remove(context.TODO())
	assert.NoError(t, err)

	lister := &CloudFrontDistributionLister{
		mockSvc: mockSvc,
	}
	opts := &nuke.ListerOpts{Config: &aws.Config{
		Region: "us-east-1",
	}}
	resources, err := lister.List(context.TODO(), opts)
	assert.NoError(t, err)
	assert.Len(t, resources, 0)

	mockSvc.AssertExpectations(t)
}

func Test_CloudFrontDistribution_RemoveUpdateETag(t *testing.T) {
	mockSvc := new(mockCloudFrontClient)
	mockSvc.On("GetDistributionConfig", mock.Anything, mock.Anything).Return(&cloudfront.GetDistributionConfigOutput{
		DistributionConfig: &types.DistributionConfig{Enabled: ptr.Bool(true)},
		ETag:               ptr.String("test-etag1"),
	}, nil)
	mockSvc.On("UpdateDistribution", mock.Anything, mock.Anything).Return(&cloudfront.UpdateDistributionOutput{
		ETag: ptr.String("test-etag2"),
	}, nil)
	mockSvc.On("DeleteDistribution", mock.Anything, &cloudfront.DeleteDistributionInput{
		Id:      ptr.String("test-id"),
		IfMatch: ptr.String("test-etag2"),
	}).Return(&cloudfront.DeleteDistributionOutput{}, nil)

	// Bug where etag was wrong so nothing was deleted
	mockSvc.On("ListDistributions", mock.Anything, mock.Anything).Return(&cloudfront.ListDistributionsOutput{
		DistributionList: &types.DistributionList{
			Items: []types.DistributionSummary{
				{Id: ptr.String("test-id"), ARN: ptr.String("test-arn"), Status: ptr.String("Deployed"), LastModifiedTime: ptr.Time(time.Now())},
			},
			IsTruncated: ptr.Bool(false),
		},
	}, nil)
	mockSvc.On("ListTagsForResource", mock.Anything, mock.Anything).Return(&cloudfront.ListTagsForResourceOutput{
		Tags: &types.Tags{
			Items: []types.Tag{
				{Key: ptr.String("test-key"), Value: ptr.String("test-value")},
			},
		},
	}, nil)

	r := &CloudFrontDistribution{
		svc: mockSvc,
		ID:  ptr.String("test-id"),
	}
	err := r.Remove(context.TODO())
	assert.NoError(t, err)

	lister := &CloudFrontDistributionLister{
		mockSvc: mockSvc,
	}
	opts := &nuke.ListerOpts{Config: &aws.Config{
		Region: "us-east-1",
	}}
	resources, err := lister.List(context.TODO(), opts)
	assert.NoError(t, err)
	assert.Len(t, resources, 1)

	mockSvc.AssertExpectations(t)
}
