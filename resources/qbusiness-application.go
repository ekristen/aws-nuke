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

const QBusinessApplicationResource = "QBusinessApplication"

func init() {
	registry.Register(&registry.Registration{
		Name:     QBusinessApplicationResource,
		Scope:    nuke.Account,
		Resource: &QBusinessApplication{},
		Lister:   &QBusinessApplicationLister{},
		DependsOn: []string{
			QBusinessWebExperienceResource,
			QBusinessPluginResource,
			QBusinessIndexResource,
			QBusinessRetrieverResource,
			QBusinessDataSourceResource,
		},
	})
}

type QBusinessApplicationLister struct{}

func (l *QBusinessApplicationLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := qbusiness.NewFromConfig(*opts.Config)
	var resources []resource.Resource

	paginator := qbusiness.NewListApplicationsPaginator(svc, &qbusiness.ListApplicationsInput{
		MaxResults: aws.Int32(100),
	})

	for paginator.HasMorePages() {
		resp, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		for _, app := range resp.Applications {
			resources = append(resources, &QBusinessApplication{
				svc:    svc,
				ID:     app.ApplicationId,
				Name:   app.DisplayName,
				Status: aws.String(string(app.Status)),
			})
		}
	}

	return resources, nil
}

type QBusinessApplication struct {
	svc    *qbusiness.Client
	ID     *string
	Name   *string
	Status *string
}

func (r *QBusinessApplication) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteApplication(ctx, &qbusiness.DeleteApplicationInput{
		ApplicationId: r.ID,
	})
	return err
}

func (r *QBusinessApplication) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *QBusinessApplication) String() string {
	return aws.ToString(r.ID)
}

// listQBusinessApplicationIDs is a shared helper used by child resource listers.
func listQBusinessApplicationIDs(ctx context.Context, svc *qbusiness.Client) ([]*string, error) {
	var appIDs []*string
	paginator := qbusiness.NewListApplicationsPaginator(svc, &qbusiness.ListApplicationsInput{
		MaxResults: aws.Int32(100),
	})
	for paginator.HasMorePages() {
		resp, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		for _, app := range resp.Applications {
			appIDs = append(appIDs, app.ApplicationId)
		}
	}
	return appIDs, nil
}
