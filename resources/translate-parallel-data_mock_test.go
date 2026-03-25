package resources

import (
	"context"
	"testing"
	"time"

	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/aws/aws-sdk-go-v2/service/translate"
	translatetypes "github.com/aws/aws-sdk-go-v2/service/translate/types"

	"github.com/ekristen/aws-nuke/v3/mocks/mock_translate"
)

func Test_Mock_TranslateParallelData_List(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mock_translate.NewMockTranslateAPI(ctrl)

	now := time.Now().UTC()

	mockSvc.EXPECT().ListParallelData(gomock.Any(), gomock.Any()).Return(&translate.ListParallelDataOutput{
		ParallelDataPropertiesList: []translatetypes.ParallelDataProperties{
			{
				Name:               ptr.String("my-parallel-data"),
				Arn:                ptr.String("arn:aws:translate:us-east-1:123456789012:parallel-data/my-parallel-data"),
				Status:             translatetypes.ParallelDataStatusActive,
				SourceLanguageCode: ptr.String("en"),
				Description:        ptr.String("test parallel data"),
				CreatedAt:          &now,
				LastUpdatedAt:      &now,
			},
			{
				Name:               ptr.String("deleting-data"),
				Arn:                ptr.String("arn:aws:translate:us-east-1:123456789012:parallel-data/deleting-data"),
				Status:             translatetypes.ParallelDataStatusDeleting,
				SourceLanguageCode: ptr.String("en"),
				CreatedAt:          &now,
				LastUpdatedAt:      &now,
			},
		},
	}, nil)

	lister := &TranslateParallelDataLister{
		svc: mockSvc,
	}

	resources, err := lister.List(context.TODO(), testListerOpts)
	a.Nil(err)
	a.Len(resources, 2)

	resource := resources[0].(*TranslateParallelData)
	a.Equal("my-parallel-data", *resource.Name)
	a.Equal(translatetypes.ParallelDataStatusActive, resource.Status)
}

func Test_Mock_TranslateParallelData_Filter(t *testing.T) {
	a := assert.New(t)

	cases := []struct {
		Status   translatetypes.ParallelDataStatus
		Filtered bool
	}{
		{Status: translatetypes.ParallelDataStatusActive, Filtered: false},
		{Status: translatetypes.ParallelDataStatusCreating, Filtered: false},
		{Status: translatetypes.ParallelDataStatusUpdating, Filtered: false},
		{Status: translatetypes.ParallelDataStatusDeleting, Filtered: true},
		{Status: translatetypes.ParallelDataStatusFailed, Filtered: true},
	}

	for _, c := range cases {
		t.Run(string(c.Status), func(t *testing.T) {
			r := &TranslateParallelData{
				Name:   ptr.String("test"),
				Status: c.Status,
			}

			err := r.Filter()
			if c.Filtered {
				a.NotNil(err)
			} else {
				a.Nil(err)
			}
		})
	}
}

func Test_Mock_TranslateParallelData_Remove(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mock_translate.NewMockTranslateAPI(ctrl)

	mockSvc.EXPECT().
		DeleteParallelData(gomock.Any(), gomock.Any()).
		Return(&translate.DeleteParallelDataOutput{}, nil)

	r := &TranslateParallelData{
		svc:  mockSvc,
		Name: ptr.String("my-parallel-data"),
	}

	err := r.Remove(context.TODO())
	a.Nil(err)
}

func Test_Mock_TranslateParallelData_Properties(t *testing.T) {
	a := assert.New(t)

	now := time.Now().UTC()

	r := &TranslateParallelData{
		Name:               ptr.String("my-parallel-data"),
		Arn:                ptr.String("arn:aws:translate:us-east-1:123456789012:parallel-data/my-parallel-data"),
		Status:             translatetypes.ParallelDataStatusActive,
		SourceLanguageCode: ptr.String("en"),
		Description:        ptr.String("test parallel data"),
		CreatedAt:          &now,
		LastUpdatedAt:      &now,
	}

	props := r.Properties()
	a.Equal("my-parallel-data", props.Get("Name"))
	a.Equal("arn:aws:translate:us-east-1:123456789012:parallel-data/my-parallel-data", props.Get("Arn"))
	a.Equal("ACTIVE", props.Get("Status"))
	a.Equal("en", props.Get("SourceLanguageCode"))
	a.Equal("test parallel data", props.Get("Description"))
}
