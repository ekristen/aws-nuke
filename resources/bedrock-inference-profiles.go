package resources

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	bedrocktypes "github.com/aws/aws-sdk-go-v2/service/bedrock/types"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const BedrockInferenceProfileResource = "BedrockInferenceProfile"

func init() {
	registry.Register(&registry.Registration{
		Name:     BedrockInferenceProfileResource,
		Scope:    nuke.Account,
		Resource: &BedrockInferenceProfile{},
		Lister:   &BedrockInferenceProfileLister{},
	})
}

type BedrockInferenceProfileLister struct{}

func (l *BedrockInferenceProfileLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := bedrock.NewFromConfig(*opts.Config)
	var resources []resource.Resource

	params := &bedrock.ListInferenceProfilesInput{
		MaxResults: aws.Int32(100),
	}

	paginator := bedrock.NewListInferenceProfilesPaginator(svc, params)

	for paginator.HasMorePages() {
		resp, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, profile := range resp.InferenceProfileSummaries {
			var tags map[string]string
			if profile.InferenceProfileArn != nil {
				tagsResp, err := svc.ListTagsForResource(ctx, &bedrock.ListTagsForResourceInput{
					ResourceARN: profile.InferenceProfileArn,
				})
				if err != nil {
					opts.Logger.Warnf("unable to fetch tags for inference profile: %s", *profile.InferenceProfileArn)
				} else {
					tags = make(map[string]string)
					for _, tag := range tagsResp.Tags {
						tags[*tag.Key] = *tag.Value
					}
				}
			}

			resources = append(resources, &BedrockInferenceProfile{
				svc:         svc,
				ID:          profile.InferenceProfileId,
				Name:        profile.InferenceProfileName,
				ARN:         profile.InferenceProfileArn,
				Status:      string(profile.Status),
				Type:        string(profile.Type),
				Description: profile.Description,
				CreatedAt:   profile.CreatedAt,
				UpdatedAt:   profile.UpdatedAt,
				Tags:        tags,
			})
		}
	}

	return resources, nil
}

type BedrockInferenceProfile struct {
	svc         *bedrock.Client
	ID          *string
	Name        *string
	ARN         *string
	Status      string
	Type        string
	Description *string
	CreatedAt   *time.Time
	UpdatedAt   *time.Time
	Tags        map[string]string
}

func (r *BedrockInferenceProfile) Filter() error {
	if r.Type == string(bedrocktypes.InferenceProfileTypeSystemDefined) {
		return fmt.Errorf("cannot delete system-defined inference profile")
	}
	return nil
}

func (r *BedrockInferenceProfile) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteInferenceProfile(ctx, &bedrock.DeleteInferenceProfileInput{
		InferenceProfileIdentifier: r.ID,
	})
	return err
}

func (r *BedrockInferenceProfile) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *BedrockInferenceProfile) String() string {
	return *r.Name
}
