package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/mediastore"
	"github.com/aws/aws-sdk-go/service/mediastoredata"

	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const MediaStoreDataItemsResource = "MediaStoreDataItems"

func init() {
	resource.Register(resource.Registration{
		Name:   MediaStoreDataItemsResource,
		Scope:  nuke.Account,
		Lister: &MediaStoreDataItemsLister{},
	})
}

type MediaStoreDataItemsLister struct{}

func (l *MediaStoreDataItemsLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	containerSvc := mediastore.New(opts.Session)
	svc := mediastoredata.New(opts.Session)
	svc.ClientInfo.SigningName = "mediastore"

	resources := make([]resource.Resource, 0)
	var containers []*mediastore.Container

	//List all containers
	containerParams := &mediastore.ListContainersInput{
		MaxResults: aws.Int64(100),
	}

	for {
		output, err := containerSvc.ListContainers(containerParams)
		if err != nil {
			return nil, err
		}

		for _, container := range output.Containers {
			containers = append(containers, container)
		}

		if output.NextToken == nil {
			break
		}

		containerParams.NextToken = output.NextToken
	}

	// List all Items per Container
	params := &mediastoredata.ListItemsInput{
		MaxResults: aws.Int64(100),
	}

	for _, container := range containers {
		if container.Endpoint == nil {
			continue
		}
		svc.Endpoint = *container.Endpoint
		output, err := svc.ListItems(params)
		if err != nil {
			return nil, err
		}

		for _, item := range output.Items {
			resources = append(resources, &MediaStoreDataItems{
				svc:  svc,
				path: item.Name,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type MediaStoreDataItems struct {
	svc  *mediastoredata.MediaStoreData
	path *string
}

func (f *MediaStoreDataItems) Remove(_ context.Context) error {
	_, err := f.svc.DeleteObject(&mediastoredata.DeleteObjectInput{
		Path: f.path,
	})

	return err
}

func (f *MediaStoreDataItems) String() string {
	return *f.path
}
