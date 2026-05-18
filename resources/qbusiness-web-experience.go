package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"               //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/qbusiness" //nolint:staticcheck

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

func (l *QBusinessWebExperienceLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := qbusiness.New(opts.Session)
	resources := make([]resource.Resource, 0)

	apps, err := listQBusinessApplications(svc)
	if err != nil {
		return nil, err
	}

	for _, appID := range apps {
		params := &qbusiness.ListWebExperiencesInput{
			ApplicationId: appID,
			MaxResults:    aws.Int64(100),
		}
		for {
			resp, err := svc.ListWebExperiences(params)
			if err != nil {
				return nil, err
			}
			for _, we := range resp.WebExperiences {
				resources = append(resources, &QBusinessWebExperience{
					svc:           svc,
					ApplicationID: appID,
					ID:            we.WebExperienceId,
					Status:        we.Status,
				})
			}
			if resp.NextToken == nil {
				break
			}
			params.NextToken = resp.NextToken
		}
	}

	return resources, nil
}

type QBusinessWebExperience struct {
	svc           *qbusiness.QBusiness
	ApplicationID *string
	ID            *string
	Status        *string
}

func (r *QBusinessWebExperience) Remove(_ context.Context) error {
	_, err := r.svc.DeleteWebExperience(&qbusiness.DeleteWebExperienceInput{
		ApplicationId:   r.ApplicationID,
		WebExperienceId: r.ID,
	})
	return err
}

func (r *QBusinessWebExperience) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *QBusinessWebExperience) String() string {
	return aws.StringValue(r.ID)
}
