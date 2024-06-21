package resources

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const SageMakerSpaceResource = "SageMakerSpace"

func init() {
	registry.Register(&registry.Registration{
		Name:   SageMakerSpaceResource,
		Scope:  nuke.Account,
		Lister: &SageMakerSpaceLister{},
	})
}

type SageMakerSpaceLister struct{}

func (l *SageMakerSpaceLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := sagemaker.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &sagemaker.ListSpacesInput{
		MaxResults: aws.Int64(30),
	}

	for {
		resp, err := svc.ListSpaces(params)
		if err != nil {
			return nil, err
		}

		for _, space := range resp.Spaces {
			resources = append(resources, &SageMakerSpace{
				svc:              svc,
				domainID:         space.DomainId,
				spaceDisplayName: space.SpaceDisplayName,
				spaceName:        space.SpaceName,
				status:           space.Status,
				lastModifiedTime: space.LastModifiedTime,
			})
		}

		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return resources, nil
}

type SageMakerSpace struct {
	svc              *sagemaker.SageMaker
	domainID         *string
	spaceDisplayName *string
	spaceName        *string
	status           *string
	lastModifiedTime *time.Time
}

func (f *SageMakerSpace) Remove(_ context.Context) error {
	_, err := f.svc.DeleteSpace(&sagemaker.DeleteSpaceInput{
		DomainId:  f.domainID,
		SpaceName: f.spaceName,
	})

	return err
}

func (f *SageMakerSpace) String() string {
	return *f.spaceName
}

func (f *SageMakerSpace) Properties() types.Properties {
	properties := types.NewProperties()
	properties.
		Set("DomainID", f.domainID).
		Set("SpaceDisplayName", f.spaceDisplayName).
		Set("SpaceName", f.spaceName).
		Set("Status", f.status).
		Set("LastModifiedTime", f.lastModifiedTime)
	return properties
}
