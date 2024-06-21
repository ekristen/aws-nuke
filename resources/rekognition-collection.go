package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rekognition"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const RekognitionCollectionResource = "RekognitionCollection"

func init() {
	registry.Register(&registry.Registration{
		Name:   RekognitionCollectionResource,
		Scope:  nuke.Account,
		Lister: &RekognitionCollectionLister{},
	})
}

type RekognitionCollectionLister struct{}

func (l *RekognitionCollectionLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := rekognition.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &rekognition.ListCollectionsInput{
		MaxResults: aws.Int64(100),
	}

	for {
		output, err := svc.ListCollections(params)
		if err != nil {
			return nil, err
		}

		for _, collection := range output.CollectionIds {
			resources = append(resources, &RekognitionCollection{
				svc: svc,
				id:  collection,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type RekognitionCollection struct {
	svc *rekognition.Rekognition
	id  *string
}

func (f *RekognitionCollection) Remove(_ context.Context) error {
	_, err := f.svc.DeleteCollection(&rekognition.DeleteCollectionInput{
		CollectionId: f.id,
	})

	return err
}

func (f *RekognitionCollection) String() string {
	return *f.id
}
