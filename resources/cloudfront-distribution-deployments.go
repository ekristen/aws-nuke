package resources

import (
	"context"
	"fmt"

	"github.com/gotidy/ptr"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudfront"

	"github.com/ekristen/aws-nuke/pkg/nuke"
	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
)

type CloudFrontDistributionDeployment struct {
	svc                *cloudfront.CloudFront
	distributionID     *string
	eTag               *string
	distributionConfig *cloudfront.DistributionConfig
	status             string
}

const CloudFrontDistributionDeploymentResource = "CloudFrontDistributionDeployment"

func init() {
	registry.Register(&registry.Registration{
		Name:   CloudFrontDistributionDeploymentResource,
		Scope:  nuke.Account,
		Lister: &CloudFrontDistributionDeploymentLister{},
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
		params := &cloudfront.GetDistributionInput{
			Id: distribution.Id,
		}
		resp, err := svc.GetDistribution(params)
		if err != nil {
			return nil, err
		}
		resources = append(resources, &CloudFrontDistributionDeployment{
			svc:                svc,
			distributionID:     resp.Distribution.Id,
			eTag:               resp.ETag,
			distributionConfig: resp.Distribution.DistributionConfig,
			status:             ptr.ToString(resp.Distribution.Status),
		})
	}

	return resources, nil
}

func (f *CloudFrontDistributionDeployment) Remove(_ context.Context) error {
	f.distributionConfig.Enabled = aws.Bool(false)

	_, err := f.svc.UpdateDistribution(&cloudfront.UpdateDistributionInput{
		Id:                 f.distributionID,
		DistributionConfig: f.distributionConfig,
		IfMatch:            f.eTag,
	})

	return err
}

func (f *CloudFrontDistributionDeployment) Filter() error {
	if !ptr.ToBool(f.distributionConfig.Enabled) && f.status != "InProgress" {
		return fmt.Errorf("already disabled")
	}
	return nil
}

func (f *CloudFrontDistributionDeployment) String() string {
	return *f.distributionID
}
