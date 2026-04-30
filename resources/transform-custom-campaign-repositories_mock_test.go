package resources

import (
	"context"
	"testing"
	"time"

	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func Test_Mock_TransformCustomCampaignRepository_List(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := NewMockTransformCustomAPI(ctrl)

	now := time.Now().UTC()

	// First, list campaigns
	mockSvc.EXPECT().ListCampaigns(gomock.Any(), gomock.Any()).Return(&TransformCustomListCampaignsOutput{
		Campaigns: []TransformCustomCampaignModel{
			{Name: "campaign-1"},
		},
	}, nil)

	// Then, list repositories for that campaign
	mockSvc.EXPECT().ListCampaignRepositories(gomock.Any(), gomock.Any()).Return(
		&TransformCustomListCampaignRepositoriesOutput{
			Repositories: []TransformCustomCampaignRepositoryModel{
				{
					Name:        "repo-1",
					Status:      "COMPLETED",
					LastUpdated: now,
				},
				{
					Name:        "repo-2",
					Status:      "IN_PROGRESS",
					LastUpdated: now,
				},
			},
		}, nil)

	lister := &TransformCustomCampaignRepositoryLister{
		svc: mockSvc,
	}

	resources, err := lister.List(context.TODO(), testListerOpts)
	a.Nil(err)
	a.Len(resources, 2)

	repo := resources[0].(*TransformCustomCampaignRepository)
	a.Equal("repo-1", *repo.Name)
	a.Equal("campaign-1", *repo.CampaignName)
	a.Equal("COMPLETED", *repo.Status)
}

func Test_Mock_TransformCustomCampaignRepository_Filter(t *testing.T) {
	a := assert.New(t)

	cases := []struct {
		name     string
		status   string
		filtered bool
	}{
		{name: "not-filtered/completed", status: "COMPLETED", filtered: false},
		{name: "not-filtered/not-started", status: "NOT_STARTED", filtered: false},
		{name: "not-filtered/validated", status: "VALIDATED", filtered: false},
		{name: "filtered/in-progress", status: "IN_PROGRESS", filtered: true},
		{name: "filtered/transforming", status: "TRANSFORMING", filtered: true},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			repo := &TransformCustomCampaignRepository{
				Name:   ptr.String("test-repo"),
				Status: ptr.String(c.status),
			}
			err := repo.Filter()
			if c.filtered {
				a.NotNil(err)
			} else {
				a.Nil(err)
			}
		})
	}
}

func Test_Mock_TransformCustomCampaignRepository_Remove(t *testing.T) {
	a := assert.New(t)

	// Remove is a no-op for campaign repositories
	repo := &TransformCustomCampaignRepository{
		Name:         ptr.String("test-repo"),
		CampaignName: ptr.String("campaign-1"),
		Status:       ptr.String("COMPLETED"),
	}

	err := repo.Remove(context.TODO())
	a.Nil(err)
}

func Test_Mock_TransformCustomCampaignRepository_Properties(t *testing.T) {
	a := assert.New(t)

	now := time.Now().UTC()
	repo := &TransformCustomCampaignRepository{
		Name:         ptr.String("test-repo"),
		CampaignName: ptr.String("campaign-1"),
		Status:       ptr.String("COMPLETED"),
		LastUpdated:  ptr.Time(now),
	}

	properties := repo.Properties()
	a.Equal("test-repo", properties.Get("Name"))
	a.Equal("campaign-1", properties.Get("CampaignName"))
	a.Equal("COMPLETED", properties.Get("Status"))
}
