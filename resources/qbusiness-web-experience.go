package resources

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/qbusiness"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const QBusinessWebExperienceResource = "QBusinessWebExperience"

func init() {
	registry.Register(&registry.Registration{
		Name:     QBusinessWebExperienceResource,
		Scope:    nuke.Account,
		Resource: &QBusinessWebExperience{},
		Lister:   &QBusinessWebExperienceLister{},
	})
}

type QBusinessWebExperienceLister struct{}

func (l *QBusinessWebExperienceLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := qbusiness.NewFromConfig(*opts.Config)
	var resources []resource.Resource

	appIDs, err := listQBusinessApplicationIDs(ctx, svc)
	if err != nil {
		return nil, err
	}

	for _, appID := range appIDs {
		paginator := qbusiness.NewListWebExperiencesPaginator(svc, &qbusiness.ListWebExperiencesInput{
			ApplicationId: appID,
			MaxResults:    aws.Int32(100),
		})
		for paginator.HasMorePages() {
			resp, err := paginator.NextPage(ctx)
			if err != nil {
				return nil, err
			}
			for _, we := range resp.WebExperiences {
				resources = append(resources, &QBusinessWebExperience{
					svc:           svc,
					ApplicationID: appID,
					ID:            we.WebExperienceId,
					Status:        aws.String(string(we.Status)),
				})
			}
		}
	}

	return resources, nil
}

type QBusinessWebExperience struct {
	svc           *qbusiness.Client
	ApplicationID *string
	ID            *string
	Status        *string
}

func (r *QBusinessWebExperience) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteWebExperience(ctx, &qbusiness.DeleteWebExperienceInput{
		ApplicationId:   r.ApplicationID,
		WebExperienceId: r.ID,
	})
	return err
}

func (r *QBusinessWebExperience) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *QBusinessWebExperience) String() string {
	return aws.ToString(r.ID)
}
