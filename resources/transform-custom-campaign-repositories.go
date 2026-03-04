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

const TransformCustomCampaignRepositoryResource = "TransformCustomCampaignRepository"

func init() {
	registry.Register(&registry.Registration{
		Name:     TransformCustomCampaignRepositoryResource,
		Scope:    nuke.Account,
		Resource: &TransformCustomCampaignRepository{},
		Lister:   &TransformCustomCampaignRepositoryLister{},
		DependsOn: []string{
			TransformCustomCampaignResource,
		},
	})
}

type TransformCustomCampaignRepositoryLister struct {
	svc TransformCustomAPI
}

func (l *TransformCustomCampaignRepositoryLister) List(
	ctx context.Context, o interface{},
) ([]resource.Resource, error) {
	if l.svc == nil {
		opts := o.(*nuke.ListerOpts)
		l.svc = NewTransformCustomClient(opts.Config)
	}

	// First, list all campaigns to enumerate repositories across them
	campaignParams := &TransformCustomListCampaignsInput{
		MaxResults: 100,
	}

	var campaignNames []string

	for {
		campaignResp, err := l.svc.ListCampaigns(ctx, campaignParams)
		if err != nil {
			return nil, err
		}

		for _, c := range campaignResp.Campaigns {
			campaignNames = append(campaignNames, c.Name)
		}

		if campaignResp.NextToken == "" {
			break
		}

		campaignParams.NextToken = campaignResp.NextToken
	}

	var resources []resource.Resource

	for _, campaignName := range campaignNames {
		params := &TransformCustomListCampaignRepositoriesInput{
			Name:       campaignName,
			MaxResults: 100,
		}

		for {
			resp, err := l.svc.ListCampaignRepositories(ctx, params)
			if err != nil {
				return nil, err
			}

			for _, repo := range resp.Repositories {
				resources = append(resources, &TransformCustomCampaignRepository{
					svc:          l.svc,
					Name:         ptr.String(repo.Name),
					CampaignName: ptr.String(campaignName),
					Status:       ptr.String(repo.Status),
					LastUpdated:  ptr.Time(repo.LastUpdated),
				})
			}

			if resp.NextToken == "" {
				break
			}

			params.NextToken = resp.NextToken
		}
	}

	return resources, nil
}

type TransformCustomCampaignRepository struct {
	svc          TransformCustomAPI
	Name         *string
	CampaignName *string
	Status       *string
	LastUpdated  *time.Time
}

func (r *TransformCustomCampaignRepository) Filter() error {
	status := ptr.ToString(r.Status)
	if status == "IN_PROGRESS" || status == "TRANSFORMING" {
		return fmt.Errorf("campaign repository is %s", status)
	}
	return nil
}

// Remove is a no-op; campaign repositories are cleaned up when the parent campaign is deleted.
func (r *TransformCustomCampaignRepository) Remove(_ context.Context) error {
	return nil
}

func (r *TransformCustomCampaignRepository) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *TransformCustomCampaignRepository) String() string {
	return ptr.ToString(r.Name)
}
