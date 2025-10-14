package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/cloudfront" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/awsutil"
	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const CloudFrontOriginRequestPolicyResource = "CloudFrontOriginRequestPolicy"

func init() {
	registry.Register(&registry.Registration{
		Name:     CloudFrontOriginRequestPolicyResource,
		Scope:    nuke.Account,
		Resource: &CloudFrontOriginRequestPolicy{},
		Lister:   &CloudFrontOriginRequestPolicyLister{},
	})
}

type CloudFrontOriginRequestPolicyLister struct{}

func (l *CloudFrontOriginRequestPolicyLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := cloudfront.New(opts.Session)
	resources := make([]resource.Resource, 0)
	params := &cloudfront.ListOriginRequestPoliciesInput{}

	for {
		resp, err := svc.ListOriginRequestPolicies(params)
		if err != nil {
			return nil, err
		}

		for _, item := range resp.OriginRequestPolicyList.Items {
			if *item.Type == awsutil.Custom {
				resources = append(resources, &CloudFrontOriginRequestPolicy{
					svc: svc,
					ID:  item.OriginRequestPolicy.Id,
				})
			}
		}

		if resp.OriginRequestPolicyList.NextMarker == nil {
			break
		}

		params.Marker = resp.OriginRequestPolicyList.NextMarker
	}

	return resources, nil
}

type CloudFrontOriginRequestPolicy struct {
	svc *cloudfront.CloudFront
	ID  *string
}

func (f *CloudFrontOriginRequestPolicy) Remove(_ context.Context) error {
	resp, err := f.svc.GetOriginRequestPolicy(&cloudfront.GetOriginRequestPolicyInput{
		Id: f.ID,
	})
	if err != nil {
		return err
	}

	_, err = f.svc.DeleteOriginRequestPolicy(&cloudfront.DeleteOriginRequestPolicyInput{
		Id:      f.ID,
		IfMatch: resp.ETag,
	})

	return err
}

func (f *CloudFrontOriginRequestPolicy) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("ID", f.ID)
	return properties
}
