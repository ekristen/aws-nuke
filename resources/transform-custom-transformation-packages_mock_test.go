package resources

import (
	"context"
	"testing"
	"time"

	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func Test_Mock_TransformCustomTransformationPackage_List(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := NewMockTransformCustomAPI(ctrl)

	now := time.Now().UTC()

	mockSvc.EXPECT().ListTransformationPackageMetadata(gomock.Any(), gomock.Any()).Return(
		&TransformCustomListTransformationPackageMetadataOutput{
			Items: []TransformCustomTransformationPackageModel{
				{
					Name:        "custom-pkg",
					Version:     "1.0.0",
					Description: "A custom package",
					CreatedAt:   now,
					Verified:    false,
					Owner:       "user",
				},
				{
					Name:        "aws-pkg",
					Version:     "2.0.0",
					Description: "An AWS package",
					CreatedAt:   now,
					Verified:    true,
					Owner:       "AWS",
				},
			},
		}, nil)

	lister := &TransformCustomTransformationPackageLister{
		svc: mockSvc,
	}

	resources, err := lister.List(context.TODO(), testListerOpts)
	a.Nil(err)
	a.Len(resources, 2)

	pkg := resources[0].(*TransformCustomTransformationPackage)
	a.Equal("custom-pkg", *pkg.Name)
	a.Equal("1.0.0", *pkg.Version)
}

func Test_Mock_TransformCustomTransformationPackage_Filter(t *testing.T) {
	a := assert.New(t)

	cases := []struct {
		name     string
		owner    string
		filtered bool
	}{
		{name: "not-filtered/user", owner: "user", filtered: false},
		{name: "filtered/aws", owner: "AWS", filtered: true},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			pkg := &TransformCustomTransformationPackage{
				Name:  ptr.String("test-pkg"),
				Owner: ptr.String(c.owner),
			}
			err := pkg.Filter()
			if c.filtered {
				a.NotNil(err)
			} else {
				a.Nil(err)
			}
		})
	}
}

func Test_Mock_TransformCustomTransformationPackage_Remove(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := NewMockTransformCustomAPI(ctrl)

	mockSvc.EXPECT().
		DeleteTransformationPackage(gomock.Any(), gomock.Any()).
		Return(&TransformCustomDeleteTransformationPackageOutput{Name: "test-pkg"}, nil)

	pkg := &TransformCustomTransformationPackage{
		svc:   mockSvc,
		Name:  ptr.String("test-pkg"),
		Owner: ptr.String("user"),
	}

	err := pkg.Remove(context.TODO())
	a.Nil(err)
}

func Test_Mock_TransformCustomTransformationPackage_Properties(t *testing.T) {
	a := assert.New(t)

	now := time.Now().UTC()
	pkg := &TransformCustomTransformationPackage{
		Name:        ptr.String("test-pkg"),
		Version:     ptr.String("1.0.0"),
		Description: ptr.String("desc"),
		CreatedAt:   ptr.Time(now),
		Verified:    ptr.Bool(true),
		Owner:       ptr.String("user"),
	}

	properties := pkg.Properties()
	a.Equal("test-pkg", properties.Get("Name"))
	a.Equal("1.0.0", properties.Get("Version"))
	a.Equal("user", properties.Get("Owner"))
}
