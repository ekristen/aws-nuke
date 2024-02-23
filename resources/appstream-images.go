package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/gotidy/ptr"

	"github.com/aws/aws-sdk-go/service/appstream"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

type AppStreamImage struct {
	svc        *appstream.AppStream
	name       *string
	visibility *string
}

const AppStreamImageResource = "AppStreamImage"

func init() {
	registry.Register(&registry.Registration{
		Name:   AppStreamImageResource,
		Scope:  nuke.Account,
		Lister: &AppStreamImageLister{},
	})
}

type AppStreamImageLister struct{}

func (l *AppStreamImageLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := appstream.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &appstream.DescribeImagesInput{}

	output, err := svc.DescribeImages(params)
	if err != nil {
		return nil, err
	}

	for _, image := range output.Images {
		resources = append(resources, &AppStreamImage{
			svc:        svc,
			name:       image.Name,
			visibility: image.Visibility,
		})
	}

	return resources, nil
}

func (f *AppStreamImage) Remove(_ context.Context) error {
	_, err := f.svc.DeleteImage(&appstream.DeleteImageInput{
		Name: f.name,
	})

	return err
}

func (f *AppStreamImage) String() string {
	return *f.name
}

func (f *AppStreamImage) Filter() error {
	if strings.EqualFold(ptr.ToString(f.visibility), "PUBLIC") {
		return fmt.Errorf("cannot delete public AWS images")
	}
	return nil
}
