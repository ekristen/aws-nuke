package resources

import (
	"context"

	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudfront"

	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const CloudFrontDistributionResource = "CloudFrontDistribution"

func init() {
	resource.Register(resource.Registration{
		Name:   CloudFrontDistributionResource,
		Scope:  nuke.Account,
		Lister: &CloudFrontDistributionLister{},
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
				status:           item.Status,
				lastModifiedTime: item.LastModifiedTime,
				tags:             tagResp.Tags.Items,
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
	status           *string
	lastModifiedTime *time.Time
	tags             []*cloudfront.Tag
}

func (f *CloudFrontDistribution) Properties() types.Properties {
	properties := types.NewProperties().
		Set("LastModifiedTime", f.lastModifiedTime.Format(time.RFC3339))

	for _, t := range f.tags {
		properties.SetTag(t.Key, t.Value)
	}
	return properties
}

func (f *CloudFrontDistribution) Remove(_ context.Context) error {
	// Get Existing eTag
	resp, err := f.svc.GetDistributionConfig(&cloudfront.GetDistributionConfigInput{
		Id: f.ID,
	})
	if err != nil {
		return err
	}

	if *resp.DistributionConfig.Enabled {
		*resp.DistributionConfig.Enabled = false
		_, err := f.svc.UpdateDistribution(&cloudfront.UpdateDistributionInput{
			Id:                 f.ID,
			DistributionConfig: resp.DistributionConfig,
			IfMatch:            resp.ETag,
		})
		if err != nil {
			return err
		}
	}

	_, err = f.svc.DeleteDistribution(&cloudfront.DeleteDistributionInput{
		Id:      f.ID,
		IfMatch: resp.ETag,
	})

	return err
}

func (f *CloudFrontDistribution) String() string {
	return *f.ID
}
