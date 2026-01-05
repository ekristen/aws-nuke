package resources

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"

	"github.com/aws/aws-sdk-go/service/textract" //nolint:staticcheck

	"github.com/ekristen/aws-nuke/v3/mocks/mock_textractiface"
)

func Test_Mock_TextractAdapter_List(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTextract := mock_textractiface.NewMockTextractAPI(ctrl)

	now := time.Now().UTC()

	// First, expect ListAdapters to be called
	mockTextract.EXPECT().ListAdapters(&textract.ListAdaptersInput{
		MaxResults: ptr.Int64(100),
	}).Return(&textract.ListAdaptersOutput{
		Adapters: []*textract.AdapterOverview{
			{
				AdapterId:    ptr.String("adapter-1"),
				AdapterName:  ptr.String("test-adapter"),
				CreationTime: ptr.Time(now),
				FeatureTypes: []*string{ptr.String("QUERIES")},
			},
		},
	}, nil)

	// Then, expect GetAdapter for each adapter to get tags
	mockTextract.EXPECT().GetAdapter(&textract.GetAdapterInput{
		AdapterId: ptr.String("adapter-1"),
	}).Return(&textract.GetAdapterOutput{
		AdapterId:   ptr.String("adapter-1"),
		AdapterName: ptr.String("test-adapter"),
		AutoUpdate:  ptr.String(textract.AutoUpdateEnabled),
		Description: ptr.String("Test adapter description"),
		Tags: map[string]*string{
			"Environment": ptr.String("test"),
		},
	}, nil)

	lister := &TextractAdapterLister{
		mockSvc: mockTextract,
	}

	resources, err := lister.List(context.TODO(), testListerOpts)
	a.Nil(err)
	a.Len(resources, 1)

	resource := resources[0].(*TextractAdapter)
	a.Equal("adapter-1", *resource.AdapterID)
	a.Equal("test-adapter", *resource.AdapterName)
	a.Equal(textract.AutoUpdateEnabled, *resource.AutoUpdate)
	a.Equal("test", resource.Tags["Environment"])
}

func Test_Mock_TextractAdapter_Remove(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTextract := mock_textractiface.NewMockTextractAPI(ctrl)

	mockTextract.EXPECT().DeleteAdapter(&textract.DeleteAdapterInput{
		AdapterId: ptr.String("adapter-1"),
	}).Return(&textract.DeleteAdapterOutput{}, nil)

	resource := &TextractAdapter{
		svc:       mockTextract,
		AdapterID: ptr.String("adapter-1"),
	}

	err := resource.Remove(context.TODO())
	a.Nil(err)
}

func Test_Mock_TextractAdapter_Properties(t *testing.T) {
	a := assert.New(t)

	now := time.Now().UTC()

	resource := &TextractAdapter{
		AdapterID:    ptr.String("adapter-1"),
		AdapterName:  ptr.String("test-adapter"),
		AutoUpdate:   ptr.String(textract.AutoUpdateEnabled),
		Description:  ptr.String("Test description"),
		CreationTime: ptr.Time(now),
		Tags: map[string]string{
			"Environment": "test",
		},
	}

	props := resource.Properties()
	a.Equal("adapter-1", props.Get("AdapterID"))
	a.Equal("test-adapter", props.Get("AdapterName"))
	a.Equal(textract.AutoUpdateEnabled, props.Get("AutoUpdate"))
	a.Equal("test", props.Get("tag:Environment"))

	a.Equal("adapter-1", resource.String())
}
