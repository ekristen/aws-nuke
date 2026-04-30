package resources

import (
	"context"
	"testing"

	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func Test_Mock_TransformCustomKnowledgeItem_List(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := NewMockTransformCustomAPI(ctrl)

	// First, list transformation packages
	mockSvc.EXPECT().ListTransformationPackageMetadata(gomock.Any(), gomock.Any()).Return(
		&TransformCustomListTransformationPackageMetadataOutput{
			Items: []TransformCustomTransformationPackageModel{
				{Name: "pkg-1"},
			},
		}, nil)

	// Then, list knowledge items for that package
	mockSvc.EXPECT().ListKnowledgeItems(gomock.Any(), gomock.Any()).Return(
		&TransformCustomListKnowledgeItemsOutput{
			KnowledgeItems: []TransformCustomKnowledgeItemModel{
				{
					ID:                        "ki-001",
					TransformationPackageName: "pkg-1",
					Title:                     "Test Rule",
					Status:                    "ENABLED",
				},
				{
					ID:                        "ki-002",
					TransformationPackageName: "pkg-1",
					Title:                     "Disabled Rule",
					Status:                    "DISABLED",
				},
			},
		}, nil)

	lister := &TransformCustomKnowledgeItemLister{
		svc: mockSvc,
	}

	resources, err := lister.List(context.TODO(), testListerOpts)
	a.Nil(err)
	a.Len(resources, 2)

	ki := resources[0].(*TransformCustomKnowledgeItem)
	a.Equal("ki-001", *ki.ID)
	a.Equal("pkg-1", *ki.TransformationPackageName)
	a.Equal("ENABLED", *ki.Status)
}

func Test_Mock_TransformCustomKnowledgeItem_Remove(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := NewMockTransformCustomAPI(ctrl)

	mockSvc.EXPECT().
		DeleteKnowledgeItem(gomock.Any(), gomock.Any()).
		Return(&TransformCustomDeleteKnowledgeItemOutput{ID: "ki-001"}, nil)

	ki := &TransformCustomKnowledgeItem{
		svc:                       mockSvc,
		ID:                        ptr.String("ki-001"),
		TransformationPackageName: ptr.String("pkg-1"),
	}

	err := ki.Remove(context.TODO())
	a.Nil(err)
}

func Test_Mock_TransformCustomKnowledgeItem_Properties(t *testing.T) {
	a := assert.New(t)

	ki := &TransformCustomKnowledgeItem{
		ID:                        ptr.String("ki-001"),
		TransformationPackageName: ptr.String("pkg-1"),
		Title:                     ptr.String("Test Rule"),
		Status:                    ptr.String("ENABLED"),
	}

	properties := ki.Properties()
	a.Equal("ki-001", properties.Get("ID"))
	a.Equal("pkg-1", properties.Get("TransformationPackageName"))
	a.Equal("Test Rule", properties.Get("Title"))
	a.Equal("ENABLED", properties.Get("Status"))
}
