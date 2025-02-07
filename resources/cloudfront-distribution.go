package resources

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	rtypes "github.com/aws/aws-sdk-go-v2/service/cloudfront/types"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const CloudFrontDistributionResource = "CloudFrontDistribution"

func init() {
	registry.Register(&registry.Registration{
		Name:     CloudFrontDistributionResource,
		Scope:    nuke.Account,
		Resource: &CloudFrontDistribution{},
		Lister:   &CloudFrontDistributionLister{},
		DependsOn: []string{
			CloudFrontDistributionDeploymentResource,
		},
	})
}

type CloudFrontDistributionLister struct {
	mockSvc CloudFrontClient
}

func (l *CloudFrontDistributionLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	var svc CloudFrontClient
	if l.mockSvc != nil {
		svc = l.mockSvc
	} else {
		svc = cloudfront.NewFromConfig(*opts.Config)
	}

	resources := make([]resource.Resource, 0)

	params := &cloudfront.ListDistributionsInput{
		MaxItems: aws.Int32(25),
	}

	for {
		resp, err := svc.ListDistributions(ctx, params)
		if err != nil {
			return nil, err
		}
		for _, item := range resp.DistributionList.Items {
			tagResp, err := svc.ListTagsForResource(ctx, &cloudfront.ListTagsForResourceInput{
				Resource: item.ARN,
			})
			if err != nil {
				return nil, err
			}

			resources = append(resources, &CloudFrontDistribution{
				svc:              svc,
				ID:               item.Id,
				Status:           item.Status,
				LastModifiedTime: item.LastModifiedTime,
				Tags:             tagResp.Tags.Items,
			})
		}

		if !*resp.DistributionList.IsTruncated {
			break
		}

		params.Marker = resp.DistributionList.NextMarker
	}

	return resources, nil
}

type CloudFrontDistribution struct {
	svc              CloudFrontClient
	ID               *string
	Status           *string
	LastModifiedTime *time.Time
	Tags             []rtypes.Tag
}

func (r *CloudFrontDistribution) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *CloudFrontDistribution) Remove(ctx context.Context) error {
	// Get Existing eTag
	resp, err := r.svc.GetDistributionConfig(ctx, &cloudfront.GetDistributionConfigInput{
		Id: r.ID,
	})
	if err != nil {
		return err
	}

	if *resp.DistributionConfig.Enabled {
		*resp.DistributionConfig.Enabled = false
		_, err := r.svc.UpdateDistribution(ctx, &cloudfront.UpdateDistributionInput{
			Id:                 r.ID,
			DistributionConfig: resp.DistributionConfig,
			IfMatch:            resp.ETag,
		})
		if err != nil {
			return err
		}
	}

	_, err = r.svc.DeleteDistribution(ctx, &cloudfront.DeleteDistributionInput{
		Id:      r.ID,
		IfMatch: resp.ETag,
	})

	return err
}

func (r *CloudFrontDistribution) String() string {
	return *r.ID
}
