package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/mediapackage"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const MediaPackageOriginEndpointResource = "MediaPackageOriginEndpoint"

func init() {
	registry.Register(&registry.Registration{
		Name:     MediaPackageOriginEndpointResource,
		Scope:    nuke.Account,
		Resource: &MediaPackageOriginEndpoint{},
		Lister:   &MediaPackageOriginEndpointLister{},
	})
}

type MediaPackageOriginEndpointLister struct{}

func (l *MediaPackageOriginEndpointLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := mediapackage.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &mediapackage.ListOriginEndpointsInput{
		MaxResults: aws.Int64(50),
	}

	for {
		output, err := svc.ListOriginEndpoints(params)
		if err != nil {
			return nil, err
		}

		for _, originEndpoint := range output.OriginEndpoints {
			resources = append(resources, &MediaPackageOriginEndpoint{
				svc: svc,
				ID:  originEndpoint.Id,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type MediaPackageOriginEndpoint struct {
	svc *mediapackage.MediaPackage
	ID  *string
}

func (f *MediaPackageOriginEndpoint) Remove(_ context.Context) error {
	_, err := f.svc.DeleteOriginEndpoint(&mediapackage.DeleteOriginEndpointInput{
		Id: f.ID,
	})

	return err
}

func (f *MediaPackageOriginEndpoint) String() string {
	return *f.ID
}
