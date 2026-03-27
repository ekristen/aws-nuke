package resources

import (
	"context"
	"testing"
	"time"

	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func Test_Mock_TransformCustomCampaign_List(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := NewMockTransformCustomAPI(ctrl)

	now := time.Now().UTC()

	mockSvc.EXPECT().ListCampaigns(gomock.Any(), gomock.Any()).Return(&TransformCustomListCampaignsOutput{
		Campaigns: []TransformCustomCampaignModel{
			{
				Name:                      "campaign-open",
				Description:               "An open campaign",
				Status:                    "OPEN",
				TransformationPackageName: "pkg-1",
				CreatedAt:                 now,
				LastUpdated:               now,
			},
			{
				Name:                      "campaign-closed",
				Description:               "A closed campaign",
				Status:                    "CLOSED",
				TransformationPackageName: "pkg-2",
				CreatedAt:                 now,
				LastUpdated:               now,
			},
		},
	}, nil)

	lister := &TransformCustomCampaignLister{
		svc: mockSvc,
	}

	resources, err := lister.List(context.TODO(), testListerOpts)
	a.Nil(err)
	a.Len(resources, 2)

	campaign := resources[0].(*TransformCustomCampaign)
	a.Equal("campaign-open", *campaign.Name)
	a.Equal("OPEN", *campaign.Status)
}

func Test_Mock_TransformCustomCampaign_Filter(t *testing.T) {
	a := assert.New(t)

	cases := []struct {
		name     string
		status   string
		filtered bool
	}{
		{name: "not-filtered/open", status: "OPEN", filtered: false},
		{name: "filtered/closed", status: "CLOSED", filtered: true},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			campaign := &TransformCustomCampaign{
				Name:   ptr.String("test-campaign"),
				Status: ptr.String(c.status),
			}
			err := campaign.Filter()
			if c.filtered {
				a.NotNil(err)
			} else {
				a.Nil(err)
			}
		})
	}
}

func Test_Mock_TransformCustomCampaign_Remove(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := NewMockTransformCustomAPI(ctrl)

	mockSvc.EXPECT().
		DeleteCampaign(gomock.Any(), gomock.Any()).
		Return(&TransformCustomDeleteCampaignOutput{Name: "test-campaign"}, nil)

	campaign := &TransformCustomCampaign{
		svc:    mockSvc,
		Name:   ptr.String("test-campaign"),
		Status: ptr.String("OPEN"),
	}

	err := campaign.Remove(context.TODO())
	a.Nil(err)
}

func Test_Mock_TransformCustomCampaign_Properties(t *testing.T) {
	a := assert.New(t)

	now := time.Now().UTC()
	campaign := &TransformCustomCampaign{
		Name:                      ptr.String("test-campaign"),
		Description:               ptr.String("desc"),
		Status:                    ptr.String("OPEN"),
		TransformationPackageName: ptr.String("pkg-1"),
		CreatedAt:                 ptr.Time(now),
		LastUpdated:               ptr.Time(now),
	}

	properties := campaign.Properties()
	a.Equal("test-campaign", properties.Get("Name"))
	a.Equal("OPEN", properties.Get("Status"))
	a.Equal("pkg-1", properties.Get("TransformationPackageName"))
}
