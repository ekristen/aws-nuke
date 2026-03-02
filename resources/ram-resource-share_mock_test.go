package resources

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/aws/aws-sdk-go-v2/service/ram"
	ramtypes "github.com/aws/aws-sdk-go-v2/service/ram/types"

	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/v3/mocks/mock_ram"
)

var owningAccountID = "123456123456"

func Test_Mock_RAMResourceShare_List(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRAM := mock_ram.NewMockRAMAPI(ctrl)

	mockRAM.EXPECT().GetResourceShares(gomock.Any(), gomock.Any()).Return(&ram.GetResourceSharesOutput{
		ResourceShares: []ramtypes.ResourceShare{
			{
				AllowExternalPrincipals: nil,
				CreationTime:            ptr.Time(time.Now().UTC()),
				FeatureSet:              "",
				LastUpdatedTime:         ptr.Time(time.Now().UTC()),
				Name:                    ptr.String("ShareActive"),
				OwningAccountId:         &owningAccountID,
				ResourceShareArn:        ptr.String(fmt.Sprintf("arn:aws:ram:us-east-1:%s:resoure-share:ShareActive", owningAccountID)),
				Status:                  "ACTIVE",
				StatusMessage:           nil,
				Tags:                    nil,
			},
			{
				AllowExternalPrincipals: nil,
				CreationTime:            ptr.Time(time.Now().UTC()),
				FeatureSet:              "",
				LastUpdatedTime:         ptr.Time(time.Now().UTC()),
				Name:                    ptr.String("ShareDeleting"),
				OwningAccountId:         &owningAccountID,
				ResourceShareArn:        ptr.String(fmt.Sprintf("arn:aws:ram:us-east-1:%s:resoure-share:ShareDeleting", owningAccountID)),
				Status:                  "DELETING",
				StatusMessage:           nil,
				Tags:                    nil,
			},
			{
				AllowExternalPrincipals: nil,
				CreationTime:            ptr.Time(time.Now().UTC()),
				FeatureSet:              "",
				LastUpdatedTime:         ptr.Time(time.Now().UTC()),
				Name:                    ptr.String("ShareDeleted"),
				OwningAccountId:         &owningAccountID,
				ResourceShareArn:        ptr.String(fmt.Sprintf("arn:aws:ram:us-east-1:%s:resoure-share:ShareDeleted", owningAccountID)),
				Status:                  "DELETED",
				StatusMessage:           nil,
				Tags:                    nil,
			},
		},
	}, nil)

	lister := &RAMResourceShareLister{
		svc: mockRAM,
	}

	resources, err := lister.List(context.TODO(), testListerOpts)
	a.Nil(err)
	a.Len(resources, 3)

	expectedResources := []resource.Resource{
		&RAMResourceShare{
			svc:              mockRAM,
			Name:             ptr.String("ShareActive"),
			OwningAccountID:  &owningAccountID,
			ResourceShareARN: ptr.String(fmt.Sprintf("arn:aws:ram:us-east-1:%s:resoure-share:ShareActive", owningAccountID)),
			Status:           "ACTIVE",
		},
		&RAMResourceShare{
			svc:              mockRAM,
			Name:             ptr.String("ShareDeleting"),
			OwningAccountID:  &owningAccountID,
			ResourceShareARN: ptr.String(fmt.Sprintf("arn:aws:ram:us-east-1:%s:resoure-share:ShareDeleting", owningAccountID)),
			Status:           "DELETING",
		},
		&RAMResourceShare{
			svc:              mockRAM,
			Name:             ptr.String("ShareDeleted"),
			OwningAccountID:  &owningAccountID,
			ResourceShareARN: ptr.String(fmt.Sprintf("arn:aws:ram:us-east-1:%s:resoure-share:ShareDeleted", owningAccountID)),
			Status:           "DELETED",
		},
	}

	a.Equal(expectedResources, resources)
}

func Test_Mock_RAMResourceShare_Filter(t *testing.T) {
	a := assert.New(t)

	cases := []struct {
		Name             *string
		OwningAccountID  *string
		ResourceShareARN *string
		Status           ramtypes.ResourceShareStatus
		Filtered         bool
	}{
		{
			Name:             ptr.String("ShareActive"),
			OwningAccountID:  &owningAccountID,
			ResourceShareARN: ptr.String(fmt.Sprintf("arn:aws:ram:us-east-1:%s:resoure-share:ShareActive", owningAccountID)),
			Status:           ramtypes.ResourceShareStatusActive,
			// only active are not filtered since only active can be deleted
			Filtered: false,
		},
		{
			Name:             ptr.String("ShareDeleting"),
			OwningAccountID:  &owningAccountID,
			ResourceShareARN: ptr.String(fmt.Sprintf("arn:aws:ram:us-east-1:%s:resoure-share:ShareDeleting", owningAccountID)),
			Status:           ramtypes.ResourceShareStatusDeleting,
			Filtered:         true,
		},
		{
			Name:             ptr.String("ShareDeleted"),
			OwningAccountID:  &owningAccountID,
			ResourceShareARN: ptr.String(fmt.Sprintf("arn:aws:ram:us-east-1:%s:resoure-share:ShareDeleted", owningAccountID)),
			Status:           ramtypes.ResourceShareStatusDeleted,
			Filtered:         true,
		},
	}

	for _, c := range cases {
		name := c.Name
		if c.Filtered {
			name = ptr.String(fmt.Sprintf("filtered/%s", *name))
		} else {
			name = ptr.String(fmt.Sprintf("not-filtered/%s", *name))
		}

		t.Run(*name, func(t *testing.T) {
			share := &RAMResourceShare{
				Name:             c.Name,
				OwningAccountID:  c.OwningAccountID,
				ResourceShareARN: c.ResourceShareARN,
				Status:           c.Status,
			}

			err := share.Filter()
			if c.Filtered {
				a.NotNil(err)
			} else {
				a.Nil(err)
			}
		})
	}
}

func Test_Mock_RAMResourceShare_Remove(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRAM := mock_ram.NewMockRAMAPI(ctrl)

	mockRAM.EXPECT().
		DeleteResourceShare(gomock.Any(), gomock.Any()).
		Return(&ram.DeleteResourceShareOutput{}, nil)

	share := &RAMResourceShare{
		svc:              mockRAM,
		Name:             ptr.String("ShareActive"),
		OwningAccountID:  &owningAccountID,
		ResourceShareARN: ptr.String(fmt.Sprintf("arn:aws:ram:us-east-1:%s:resoure-share:ShareActive", owningAccountID)),
		Status:           "Active",
	}

	err := share.Remove(context.TODO())
	a.Nil(err)
}

func Test_Mock_RAMResourceShare_Properties(t *testing.T) {
	a := assert.New(t)

	share := &RAMResourceShare{
		Name:             ptr.String("ShareActive"),
		OwningAccountID:  &owningAccountID,
		ResourceShareARN: ptr.String(fmt.Sprintf("arn:aws:ram:us-east-1:%s:resoure-share:ShareActive", owningAccountID)),
		Status:           "ACTIVE",
	}

	properties := share.Properties()
	a.Equal("ShareActive", properties.Get("Name"))
	a.Equal(owningAccountID, properties.Get("OwningAccountID"))
	a.Equal(fmt.Sprintf("arn:aws:ram:us-east-1:%s:resoure-share:ShareActive", owningAccountID), properties.Get("ResourceShareARN"))
	a.Equal("ACTIVE", properties.Get("Status"))
}
