package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/mediastore"

	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const MediaStoreContainerResource = "MediaStoreContainer"

func init() {
	resource.Register(resource.Registration{
		Name:   MediaStoreContainerResource,
		Scope:  nuke.Account,
		Lister: &MediaStoreContainerLister{},
	})
}

type MediaStoreContainerLister struct{}

func (l *MediaStoreContainerLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := mediastore.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &mediastore.ListContainersInput{
		MaxResults: aws.Int64(100),
	}

	for {
		output, err := svc.ListContainers(params)
		if err != nil {
			return nil, err
		}

		for _, container := range output.Containers {
			resources = append(resources, &MediaStoreContainer{
				svc:  svc,
				name: container.Name,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type MediaStoreContainer struct {
	svc  *mediastore.MediaStore
	name *string
}

func (f *MediaStoreContainer) Remove(_ context.Context) error {
	_, err := f.svc.DeleteContainer(&mediastore.DeleteContainerInput{
		ContainerName: f.name,
	})

	return err
}

func (f *MediaStoreContainer) String() string {
	return *f.name
}
