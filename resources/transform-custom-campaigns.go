package resources

import (
	"context"
	"fmt"
	"time"

	"github.com/gotidy/ptr"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const TransformCustomCampaignResource = "TransformCustomCampaign"

func init() {
	registry.Register(&registry.Registration{
		Name:     TransformCustomCampaignResource,
		Scope:    nuke.Account,
		Resource: &TransformCustomCampaign{},
		Lister:   &TransformCustomCampaignLister{},
	})
}

type TransformCustomCampaignLister struct {
	svc TransformCustomAPI
}

func (l *TransformCustomCampaignLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	if l.svc == nil {
		opts := o.(*nuke.ListerOpts)
		l.svc = NewTransformCustomClient(opts.Config)
	}

	var resources []resource.Resource

	params := &TransformCustomListCampaignsInput{
		MaxResults: 100,
	}

	for {
		resp, err := l.svc.ListCampaigns(ctx, params)
		if err != nil {
			return nil, err
		}

		for _, c := range resp.Campaigns {
			resources = append(resources, &TransformCustomCampaign{
				svc:                       l.svc,
				Name:                      ptr.String(c.Name),
				Description:               ptr.String(c.Description),
				Status:                    ptr.String(c.Status),
				TransformationPackageName: ptr.String(c.TransformationPackageName),
				CreatedAt:                 ptr.Time(c.CreatedAt),
				LastUpdated:               ptr.Time(c.LastUpdated),
			})
		}

		if resp.NextToken == "" {
			break
		}

		params.NextToken = resp.NextToken
	}

	return resources, nil
}

type TransformCustomCampaign struct {
	svc                       TransformCustomAPI
	Name                      *string
	Description               *string
	Status                    *string
	TransformationPackageName *string
	CreatedAt                 *time.Time
	LastUpdated               *time.Time
}

func (r *TransformCustomCampaign) Filter() error {
	if ptr.ToString(r.Status) == "CLOSED" {
		return fmt.Errorf("campaign is CLOSED")
	}
	return nil
}

func (r *TransformCustomCampaign) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteCampaign(ctx, &TransformCustomDeleteCampaignInput{
		Name: ptr.ToString(r.Name),
	})
	return err
}

func (r *TransformCustomCampaign) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *TransformCustomCampaign) String() string {
	return ptr.ToString(r.Name)
}
