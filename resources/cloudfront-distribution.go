package resources

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudfront"

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

type CloudFrontDistributionLister struct{}

func (l *CloudFrontDistributionLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := cloudfront.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &cloudfront.ListDistributionsInput{
		MaxItems: aws.Int64(25),
	}

	for {
		resp, err := svc.ListDistributions(params)
		if err != nil {
			return nil, err
		}
		for _, item := range resp.DistributionList.Items {
			tagResp, err := svc.ListTagsForResource(
				&cloudfront.ListTagsForResourceInput{
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
	svc              *cloudfront.CloudFront
	ID               *string
	Status           *string
	LastModifiedTime *time.Time
	Tags             []*cloudfront.Tag
}

func (r *CloudFrontDistribution) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *CloudFrontDistribution) Remove(_ context.Context) error {
	// Get Existing eTag
	resp, err := r.svc.GetDistributionConfig(&cloudfront.GetDistributionConfigInput{
		Id: r.ID,
	})
	if err != nil {
		return err
	}

	if *resp.DistributionConfig.Enabled {
		*resp.DistributionConfig.Enabled = false
		_, err := r.svc.UpdateDistribution(&cloudfront.UpdateDistributionInput{
			Id:                 r.ID,
			DistributionConfig: resp.DistributionConfig,
			IfMatch:            resp.ETag,
		})
		if err != nil {
			return err
		}
	}

	_, err = r.svc.DeleteDistribution(&cloudfront.DeleteDistributionInput{
		Id:      r.ID,
		IfMatch: resp.ETag,
	})

	return err
}

func (r *CloudFrontDistribution) String() string {
	return *r.ID
}
