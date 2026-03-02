package resources

import (
	"context"
	"testing"

	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/aws/aws-sdk-go-v2/service/timestreaminfluxdb"
	influxdbtypes "github.com/aws/aws-sdk-go-v2/service/timestreaminfluxdb/types"

	"github.com/ekristen/aws-nuke/v3/mocks/mock_timestreaminfluxdb"
)

func Test_Mock_TimestreamInfluxDBDbInstance_List(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mock_timestreaminfluxdb.NewMockTimestreamInfluxDBAPI(ctrl)

	arn1 := "arn:aws:timestream-influxdb:us-east-1:123456789012:dbinstance/instance-1"
	arn2 := "arn:aws:timestream-influxdb:us-east-1:123456789012:dbinstance/instance-2"

	mockSvc.EXPECT().
		ListDbInstances(gomock.Any(), gomock.Any()).
		Return(&timestreaminfluxdb.ListDbInstancesOutput{
			Items: []influxdbtypes.DbInstanceSummary{
				{
					Id:     ptr.String("id-1"),
					Name:   ptr.String("instance-1"),
					Arn:    ptr.String(arn1),
					Status: influxdbtypes.StatusAvailable,
				},
				{
					Id:     ptr.String("id-2"),
					Name:   ptr.String("instance-2"),
					Arn:    ptr.String(arn2),
					Status: influxdbtypes.StatusDeleting,
				},
			},
		}, nil)

	mockSvc.EXPECT().
		ListTagsForResource(gomock.Any(), &timestreaminfluxdb.ListTagsForResourceInput{
			ResourceArn: ptr.String(arn1),
		}).
		Return(&timestreaminfluxdb.ListTagsForResourceOutput{
			Tags: map[string]string{"env": "test"},
		}, nil)

	mockSvc.EXPECT().
		ListTagsForResource(gomock.Any(), &timestreaminfluxdb.ListTagsForResourceInput{
			ResourceArn: ptr.String(arn2),
		}).
		Return(&timestreaminfluxdb.ListTagsForResourceOutput{
			Tags: map[string]string{},
		}, nil)

	lister := &TimestreamInfluxDBDbInstanceLister{
		svc: mockSvc,
	}

	resources, err := lister.List(context.TODO(), testListerOpts)
	a.Nil(err)
	a.Len(resources, 2)

	first := resources[0].(*TimestreamInfluxDBDbInstance)
	a.Equal("id-1", *first.ID)
	a.Equal("instance-1", *first.Name)
	a.Equal(arn1, *first.Arn)
	a.Equal("AVAILABLE", first.Status)
	a.Equal(map[string]string{"env": "test"}, first.Tags)

	second := resources[1].(*TimestreamInfluxDBDbInstance)
	a.Equal("id-2", *second.ID)
	a.Equal("instance-2", *second.Name)
	a.Equal("DELETING", second.Status)
}

func Test_Mock_TimestreamInfluxDBDbInstance_Filter(t *testing.T) {
	a := assert.New(t)

	cases := []struct {
		Name     string
		Status   string
		Filtered bool
	}{
		{
			Name:     "available",
			Status:   string(influxdbtypes.StatusAvailable),
			Filtered: false,
		},
		{
			Name:     "creating",
			Status:   string(influxdbtypes.StatusCreating),
			Filtered: false,
		},
		{
			Name:     "deleting",
			Status:   string(influxdbtypes.StatusDeleting),
			Filtered: true,
		},
		{
			Name:     "deleted",
			Status:   string(influxdbtypes.StatusDeleted),
			Filtered: true,
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			instance := &TimestreamInfluxDBDbInstance{
				ID:     ptr.String("id-1"),
				Name:   ptr.String("instance-1"),
				Status: c.Status,
			}

			err := instance.Filter()
			if c.Filtered {
				a.NotNil(err)
			} else {
				a.Nil(err)
			}
		})
	}
}

func Test_Mock_TimestreamInfluxDBDbInstance_Remove(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mock_timestreaminfluxdb.NewMockTimestreamInfluxDBAPI(ctrl)

	mockSvc.EXPECT().
		DeleteDbInstance(gomock.Any(), &timestreaminfluxdb.DeleteDbInstanceInput{
			Identifier: ptr.String("id-1"),
		}).
		Return(&timestreaminfluxdb.DeleteDbInstanceOutput{}, nil)

	instance := &TimestreamInfluxDBDbInstance{
		svc:    mockSvc,
		ID:     ptr.String("id-1"),
		Name:   ptr.String("instance-1"),
		Status: string(influxdbtypes.StatusAvailable),
	}

	err := instance.Remove(context.TODO())
	a.Nil(err)
}

func Test_Mock_TimestreamInfluxDBDbInstance_Properties(t *testing.T) {
	a := assert.New(t)

	arn := "arn:aws:timestream-influxdb:us-east-1:123456789012:dbinstance/instance-1"

	instance := &TimestreamInfluxDBDbInstance{
		ID:     ptr.String("id-1"),
		Name:   ptr.String("instance-1"),
		Arn:    ptr.String(arn),
		Status: string(influxdbtypes.StatusAvailable),
		Tags: map[string]string{
			"Environment": "production",
		},
	}

	props := instance.Properties()

	a.Equal("id-1", props.Get("ID"))
	a.Equal("instance-1", props.Get("Name"))
	a.Equal(arn, props.Get("Arn"))
	a.Equal("AVAILABLE", props.Get("Status"))
	a.Equal("production", props.Get("tag:Environment"))
}
