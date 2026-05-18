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

const QBusinessIndexResource = "QBusinessIndex"

func init() {
	registry.Register(&registry.Registration{
		Name:     QBusinessIndexResource,
		Scope:    nuke.Account,
		Resource: &QBusinessIndex{},
		Lister:   &QBusinessIndexLister{},
		DependsOn: []string{
			QBusinessDataSourceResource,
		},
	})
}

type QBusinessIndexLister struct{}

func (l *QBusinessIndexLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := qbusiness.New(opts.Session)
	resources := make([]resource.Resource, 0)

	apps, err := listQBusinessApplications(svc)
	if err != nil {
		return nil, err
	}

	for _, appID := range apps {
		params := &qbusiness.ListIndicesInput{
			ApplicationId: appID,
			MaxResults:    aws.Int64(100),
		}
		for {
			resp, err := svc.ListIndices(params)
			if err != nil {
				return nil, err
			}
			for _, idx := range resp.Indices {
				resources = append(resources, &QBusinessIndex{
					svc:           svc,
					ApplicationID: appID,
					ID:            idx.IndexId,
					Name:          idx.DisplayName,
					Status:        idx.Status,
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

type QBusinessIndex struct {
	svc           *qbusiness.QBusiness
	ApplicationID *string
	ID            *string
	Name          *string
	Status        *string
}

func (r *QBusinessIndex) Remove(_ context.Context) error {
	_, err := r.svc.DeleteIndex(&qbusiness.DeleteIndexInput{
		ApplicationId: r.ApplicationID,
		IndexId:       r.ID,
	})
	return err
}

func (r *QBusinessIndex) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *QBusinessIndex) String() string {
	return aws.StringValue(r.ID)
}
