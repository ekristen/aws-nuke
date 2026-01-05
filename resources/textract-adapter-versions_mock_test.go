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

func Test_Mock_TextractAdapterVersion_List(t *testing.T) {
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

	// Then, expect ListAdapterVersions for each adapter
	mockTextract.EXPECT().ListAdapterVersions(&textract.ListAdapterVersionsInput{
		AdapterId:  ptr.String("adapter-1"),
		MaxResults: ptr.Int64(100),
	}).Return(&textract.ListAdapterVersionsOutput{
		AdapterVersions: []*textract.AdapterVersionOverview{
			{
				AdapterId:      ptr.String("adapter-1"),
				AdapterVersion: ptr.String("1"),
				Status:         ptr.String(textract.AdapterVersionStatusActive),
				CreationTime:   ptr.Time(now),
				FeatureTypes:   []*string{ptr.String("QUERIES")},
			},
		},
	}, nil)

	lister := &TextractAdapterVersionLister{
		mockSvc: mockTextract,
	}

	resources, err := lister.List(context.TODO(), testListerOpts)
	a.Nil(err)
	a.Len(resources, 1)

	resource := resources[0].(*TextractAdapterVersion)
	a.Equal("adapter-1", *resource.AdapterID)
	a.Equal("1", *resource.AdapterVersion)
	a.Equal(textract.AdapterVersionStatusActive, *resource.Status)
}

func Test_Mock_TextractAdapterVersion_Remove(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTextract := mock_textractiface.NewMockTextractAPI(ctrl)

	mockTextract.EXPECT().DeleteAdapterVersion(&textract.DeleteAdapterVersionInput{
		AdapterId:      ptr.String("adapter-1"),
		AdapterVersion: ptr.String("1"),
	}).Return(&textract.DeleteAdapterVersionOutput{}, nil)

	resource := &TextractAdapterVersion{
		svc:            mockTextract,
		AdapterID:      ptr.String("adapter-1"),
		AdapterVersion: ptr.String("1"),
	}

	err := resource.Remove(context.TODO())
	a.Nil(err)
}

func Test_Mock_TextractAdapterVersion_Filter_CreationInProgress(t *testing.T) {
	a := assert.New(t)

	resource := &TextractAdapterVersion{
		AdapterID:      ptr.String("adapter-1"),
		AdapterVersion: ptr.String("1"),
		Status:         ptr.String(textract.AdapterVersionStatusCreationInProgress),
	}

	err := resource.Filter()
	a.NotNil(err)
	a.Contains(err.Error(), "CREATION_IN_PROGRESS")
}

func Test_Mock_TextractAdapterVersion_Filter_CreationError(t *testing.T) {
	a := assert.New(t)

	resource := &TextractAdapterVersion{
		AdapterID:      ptr.String("adapter-1"),
		AdapterVersion: ptr.String("1"),
		Status:         ptr.String(textract.AdapterVersionStatusCreationError),
	}

	err := resource.Filter()
	a.NotNil(err)
	a.Contains(err.Error(), "CREATION_ERROR")
}

func Test_Mock_TextractAdapterVersion_Filter_Active(t *testing.T) {
	a := assert.New(t)

	resource := &TextractAdapterVersion{
		AdapterID:      ptr.String("adapter-1"),
		AdapterVersion: ptr.String("1"),
		Status:         ptr.String(textract.AdapterVersionStatusActive),
	}

	err := resource.Filter()
	a.Nil(err)
}

func Test_Mock_TextractAdapterVersion_Properties(t *testing.T) {
	a := assert.New(t)

	now := time.Now().UTC()

	resource := &TextractAdapterVersion{
		AdapterID:      ptr.String("adapter-1"),
		AdapterVersion: ptr.String("1"),
		Status:         ptr.String(textract.AdapterVersionStatusActive),
		CreationTime:   ptr.Time(now),
	}

	props := resource.Properties()
	a.Equal("adapter-1", props.Get("AdapterID"))
	a.Equal("1", props.Get("AdapterVersion"))
	a.Equal(textract.AdapterVersionStatusActive, props.Get("Status"))

	a.Equal("adapter-1:1", resource.String())
}
