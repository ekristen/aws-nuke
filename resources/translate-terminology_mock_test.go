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

func Test_Mock_TranslateTerminology_List(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mock_translate.NewMockTranslateAPI(ctrl)

	now := time.Now().UTC()

	mockSvc.EXPECT().ListTerminologies(gomock.Any(), gomock.Any()).Return(&translate.ListTerminologiesOutput{
		TerminologyPropertiesList: []translatetypes.TerminologyProperties{
			{
				Name:               ptr.String("my-terminology"),
				Arn:                ptr.String("arn:aws:translate:us-east-1:123456789012:terminology/my-terminology"),
				SourceLanguageCode: ptr.String("en"),
				Description:        ptr.String("test terminology"),
				CreatedAt:          &now,
				LastUpdatedAt:      &now,
				SizeBytes:          ptr.Int32(1024),
				TermCount:          ptr.Int32(50),
			},
		},
	}, nil)

	lister := &TranslateTerminologyLister{
		svc: mockSvc,
	}

	resources, err := lister.List(context.TODO(), testListerOpts)
	a.Nil(err)
	a.Len(resources, 1)

	resource := resources[0].(*TranslateTerminology)
	a.Equal("my-terminology", *resource.Name)
	a.Equal(int32(1024), *resource.SizeBytes)
}

func Test_Mock_TranslateTerminology_Remove(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mock_translate.NewMockTranslateAPI(ctrl)

	mockSvc.EXPECT().
		DeleteTerminology(gomock.Any(), gomock.Any()).
		Return(&translate.DeleteTerminologyOutput{}, nil)

	r := &TranslateTerminology{
		svc:  mockSvc,
		Name: ptr.String("my-terminology"),
	}

	err := r.Remove(context.TODO())
	a.Nil(err)
}

func Test_Mock_TranslateTerminology_Properties(t *testing.T) {
	a := assert.New(t)

	now := time.Now().UTC()

	r := &TranslateTerminology{
		Name:               ptr.String("my-terminology"),
		Arn:                ptr.String("arn:aws:translate:us-east-1:123456789012:terminology/my-terminology"),
		SourceLanguageCode: ptr.String("en"),
		Description:        ptr.String("test terminology"),
		CreatedAt:          &now,
		LastUpdatedAt:      &now,
		SizeBytes:          ptr.Int32(1024),
		TermCount:          ptr.Int32(50),
	}

	props := r.Properties()
	a.Equal("my-terminology", props.Get("Name"))
	a.Equal("arn:aws:translate:us-east-1:123456789012:terminology/my-terminology", props.Get("Arn"))
	a.Equal("en", props.Get("SourceLanguageCode"))
	a.Equal("test terminology", props.Get("Description"))
}
