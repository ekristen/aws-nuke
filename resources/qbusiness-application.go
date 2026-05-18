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

func (l *QBusinessApplicationLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := qbusiness.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &qbusiness.ListApplicationsInput{
		MaxResults: aws.Int64(100),
	}

	for {
		resp, err := svc.ListApplications(params)
		if err != nil {
			return nil, err
		}

		for _, app := range resp.Applications {
			resources = append(resources, &QBusinessApplication{
				svc:    svc,
				ID:     app.ApplicationId,
				Name:   app.DisplayName,
				Status: app.Status,
			})
		}

		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return resources, nil
}

type QBusinessApplication struct {
	svc    *qbusiness.QBusiness
	ID     *string
	Name   *string
	Status *string
}

func (r *QBusinessApplication) Remove(_ context.Context) error {
	_, err := r.svc.DeleteApplication(&qbusiness.DeleteApplicationInput{
		ApplicationId: r.ID,
	})
	return err
}

func (r *QBusinessApplication) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *QBusinessApplication) String() string {
	return aws.StringValue(r.ID)
}

// listQBusinessApplications is a shared helper used by child resource listers.
func listQBusinessApplications(svc *qbusiness.QBusiness) ([]*string, error) {
	var appIDs []*string
	params := &qbusiness.ListApplicationsInput{MaxResults: aws.Int64(100)}
	for {
		resp, err := svc.ListApplications(params)
		if err != nil {
			return nil, err
		}
		for _, app := range resp.Applications {
			appIDs = append(appIDs, app.ApplicationId)
		}
		if resp.NextToken == nil {
			break
		}
		params.NextToken = resp.NextToken
	}
	return appIDs, nil
}
