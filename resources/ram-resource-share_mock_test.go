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

var owningAccountId = "123456123456"

func Test_Mock_RamResourceShare_List(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRam := mock_ram.NewMockRamAPI(ctrl)

	mockRam.EXPECT().GetResourceShares(gomock.Any(), gomock.Any()).Return(&ram.GetResourceSharesOutput{
		ResourceShares: []ramtypes.ResourceShare{
			{
				AllowExternalPrincipals: nil,
				CreationTime:            ptr.Time(time.Now().UTC()),
				FeatureSet:              "",
				LastUpdatedTime:         ptr.Time(time.Now().UTC()),
				Name:                    ptr.String("ShareActive"),
				OwningAccountId:         &owningAccountId,
				ResourceShareArn:        ptr.String(fmt.Sprintf("arn:aws:ram:us-east-1:%s:resoure-share:ShareActive", owningAccountId)),
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
				OwningAccountId:         &owningAccountId,
				ResourceShareArn:        ptr.String(fmt.Sprintf("arn:aws:ram:us-east-1:%s:resoure-share:ShareDeleting", owningAccountId)),
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
				OwningAccountId:         &owningAccountId,
				ResourceShareArn:        ptr.String(fmt.Sprintf("arn:aws:ram:us-east-1:%s:resoure-share:ShareDeleted", owningAccountId)),
				Status:                  "DELETED",
				StatusMessage:           nil,
				Tags:                    nil,
			},
		},
	}, nil)

	lister := &RamResourceShareLister{
		svc: mockRam,
	}

	resources, err := lister.List(context.TODO(), testListerOpts)
	a.Nil(err)
	a.Len(resources, 3)

	expectedResources := []resource.Resource{
		&RamResourceShare{
			svc:              mockRam,
			Name:             ptr.String("ShareActive"),
			OwningAccountId:  &owningAccountId,
			ResourceShareArn: ptr.String(fmt.Sprintf("arn:aws:ram:us-east-1:%s:resoure-share:ShareActive", owningAccountId)),
			Status:           "ACTIVE",
		},
		&RamResourceShare{
			svc:              mockRam,
			Name:             ptr.String("ShareDeleting"),
			OwningAccountId:  &owningAccountId,
			ResourceShareArn: ptr.String(fmt.Sprintf("arn:aws:ram:us-east-1:%s:resoure-share:ShareDeleting", owningAccountId)),
			Status:           "DELETING",
		},
		&RamResourceShare{
			svc:              mockRam,
			Name:             ptr.String("ShareDeleted"),
			OwningAccountId:  &owningAccountId,
			ResourceShareArn: ptr.String(fmt.Sprintf("arn:aws:ram:us-east-1:%s:resoure-share:ShareDeleted", owningAccountId)),
			Status:           "DELETED",
		},
	}

	a.Equal(expectedResources, resources)
}

func Test_Mock_RamResourceShare_Filter(t *testing.T) {
	a := assert.New(t)

	cases := []struct {
		Name             *string
		OwningAccountId  *string
		ResourceShareArn *string
		Status           ramtypes.ResourceShareStatus
		Filtered         bool
	}{
		{
			Name:             ptr.String("ShareActive"),
			OwningAccountId:  &owningAccountId,
			ResourceShareArn: ptr.String(fmt.Sprintf("arn:aws:ram:us-east-1:%s:resoure-share:ShareActive", owningAccountId)),
			Status:           ramtypes.ResourceShareStatusActive,
			// only active are not filtered since only active can be deleted
			Filtered: false,
		},
		{
			Name:             ptr.String("ShareDeleting"),
			OwningAccountId:  &owningAccountId,
			ResourceShareArn: ptr.String(fmt.Sprintf("arn:aws:ram:us-east-1:%s:resoure-share:ShareDeleting", owningAccountId)),
			Status:           ramtypes.ResourceShareStatusDeleting,
			Filtered:         true,
		},
		{
			Name:             ptr.String("ShareDeleted"),
			OwningAccountId:  &owningAccountId,
			ResourceShareArn: ptr.String(fmt.Sprintf("arn:aws:ram:us-east-1:%s:resoure-share:ShareDeleted", owningAccountId)),
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
			share := &RamResourceShare{
				Name:             c.Name,
				OwningAccountId:  c.OwningAccountId,
				ResourceShareArn: c.ResourceShareArn,
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

func Test_Mock_RamResourceShare_Remove(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRam := mock_ram.NewMockRamAPI(ctrl)

	mockRam.EXPECT().
		DeleteResourceShare(gomock.Any(), gomock.Any()).
		Return(&ram.DeleteResourceShareOutput{}, nil)

	share := &RamResourceShare{
		svc:              mockRam,
		Name:             ptr.String("ShareActive"),
		OwningAccountId:  &owningAccountId,
		ResourceShareArn: ptr.String(fmt.Sprintf("arn:aws:ram:us-east-1:%s:resoure-share:ShareActive", owningAccountId)),
		Status:           "Active",
	}

	err := share.Remove(context.TODO())
	a.Nil(err)
}

func Test_Mock_RamResourceShare_Properties(t *testing.T) {
	a := assert.New(t)

	share := &RamResourceShare{
		Name:             ptr.String("ShareActive"),
		OwningAccountId:  &owningAccountId,
		ResourceShareArn: ptr.String(fmt.Sprintf("arn:aws:ram:us-east-1:%s:resoure-share:ShareActive", owningAccountId)),
		Status:           "ACTIVE",
	}

	properties := share.Properties()
	a.Equal("ShareActive", properties.Get("Name"))
	a.Equal(owningAccountId, properties.Get("OwningAccountId"))
	a.Equal(fmt.Sprintf("arn:aws:ram:us-east-1:%s:resoure-share:ShareActive", owningAccountId), properties.Get("ResourceShareArn"))
	a.Equal("ACTIVE", properties.Get("Status"))
}
