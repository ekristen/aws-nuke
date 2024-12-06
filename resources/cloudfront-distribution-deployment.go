package resources

import (
	"context"
	"fmt"

	"github.com/gotidy/ptr"
	"github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudfront"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const CloudFrontDistributionDeploymentResource = "CloudFrontDistributionDeployment"

func init() {
	registry.Register(&registry.Registration{
		Name:     CloudFrontDistributionDeploymentResource,
		Scope:    nuke.Account,
		Resource: &CloudFrontDistributionDeployment{},
		Lister:   &CloudFrontDistributionDeploymentLister{},
	})
}

type CloudFrontDistributionDeploymentLister struct{}

func (l *CloudFrontDistributionDeploymentLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := cloudfront.New(opts.Session)
	resources := make([]resource.Resource, 0)
	var distributions []*cloudfront.DistributionSummary

	params := &cloudfront.ListDistributionsInput{
		MaxItems: aws.Int64(25),
	}

	for {
		resp, err := svc.ListDistributions(params)
		if err != nil {
			return nil, err
		}

		distributions = append(distributions, resp.DistributionList.Items...)

		if !ptr.ToBool(resp.DistributionList.IsTruncated) {
			break
		}

		params.Marker = resp.DistributionList.NextMarker
	}

	for _, distribution := range distributions {
		resp, err := svc.GetDistribution(&cloudfront.GetDistributionInput{
			Id: distribution.Id,
		})
		if err != nil {
			logrus.WithError(err).Error("unable to get distribution, skipping")
			continue
		}

		resources = append(resources, &CloudFrontDistributionDeployment{
			svc:                svc,
			ID:                 resp.Distribution.Id,
			eTag:               resp.ETag,
			distributionConfig: resp.Distribution.DistributionConfig,
			Status:             resp.Distribution.Status,
		})
	}

	return resources, nil
}

type CloudFrontDistributionDeployment struct {
	svc                *cloudfront.CloudFront
	ID                 *string
	Status             *string
	eTag               *string
	distributionConfig *cloudfront.DistributionConfig
}

func (r *CloudFrontDistributionDeployment) Remove(_ context.Context) error {
	r.distributionConfig.Enabled = aws.Bool(false)

	_, err := r.svc.UpdateDistribution(&cloudfront.UpdateDistributionInput{
		Id:                 r.ID,
		DistributionConfig: r.distributionConfig,
		IfMatch:            r.eTag,
	})

	return err
}

func (r *CloudFrontDistributionDeployment) Filter() error {
	if !ptr.ToBool(r.distributionConfig.Enabled) && ptr.ToString(r.Status) != "InProgress" {
		return fmt.Errorf("already disabled")
	}
	return nil
}

func (r *CloudFrontDistributionDeployment) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *CloudFrontDistributionDeployment) String() string {
	return *r.ID
}
