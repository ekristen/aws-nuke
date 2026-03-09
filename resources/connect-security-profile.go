package resources

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/connect"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const ConnectSecurityProfileResource = "ConnectSecurityProfile"

func init() {
	registry.Register(&registry.Registration{
		Name:     ConnectSecurityProfileResource,
		Scope:    nuke.Account,
		Resource: &ConnectSecurityProfile{},
		Lister:   &ConnectSecurityProfileLister{},
	})
}

type ConnectSecurityProfileLister struct{}

func (l *ConnectSecurityProfileLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := connect.NewFromConfig(*opts.Config)
	var resources []resource.Resource

	instances, err := listConnectInstances(ctx, svc)
	if err != nil {
		return nil, err
	}

	for _, instance := range instances {
		paginator := connect.NewListSecurityProfilesPaginator(svc, &connect.ListSecurityProfilesInput{
			InstanceId: instance.Id,
		})

		for paginator.HasMorePages() {
			resp, err := paginator.NextPage(ctx)
			if err != nil {
				return nil, err
			}

			for _, profile := range resp.SecurityProfileSummaryList {
				var tags map[string]string
				if profile.Arn != nil {
					tagsResp, err := svc.ListTagsForResource(ctx, &connect.ListTagsForResourceInput{
						ResourceArn: profile.Arn,
					})
					if err != nil {
						opts.Logger.Warnf("unable to fetch tags for connect security profile: %s", *profile.Arn)
					} else {
						tags = tagsResp.Tags
					}
				}

				resources = append(resources, &ConnectSecurityProfile{
					svc:        svc,
					InstanceID: instance.Id,
					ProfileID:  profile.Id,
					Name:       profile.Name,
					ARN:        profile.Arn,
					Tags:       tags,
				})
			}
		}
	}

	return resources, nil
}

type ConnectSecurityProfile struct {
	svc        *connect.Client
	InstanceID *string
	ProfileID  *string
	Name       *string
	ARN        *string
	Tags       map[string]string
}

var connectBuiltInSecurityProfiles = map[string]struct{}{
	"Admin":             {},
	"Agent":             {},
	"CallCenterManager": {},
	"QualityAnalyst":    {},
}

func (r *ConnectSecurityProfile) Filter() error {
	if r.Name != nil {
		if _, ok := connectBuiltInSecurityProfiles[*r.Name]; ok {
			return fmt.Errorf("cannot delete built-in security profile: %s", *r.Name)
		}
	}
	return nil
}

func (r *ConnectSecurityProfile) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteSecurityProfile(ctx, &connect.DeleteSecurityProfileInput{
		InstanceId:        r.InstanceID,
		SecurityProfileId: r.ProfileID,
	})
	return err
}

func (r *ConnectSecurityProfile) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *ConnectSecurityProfile) String() string {
	return *r.Name
}
