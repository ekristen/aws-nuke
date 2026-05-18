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

const QBusinessDataSourceResource = "QBusinessDataSource"

func init() {
	registry.Register(&registry.Registration{
		Name:     QBusinessDataSourceResource,
		Scope:    nuke.Account,
		Resource: &QBusinessDataSource{},
		Lister:   &QBusinessDataSourceLister{},
	})
}

type QBusinessDataSourceLister struct{}

func (l *QBusinessDataSourceLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := qbusiness.New(opts.Session)
	resources := make([]resource.Resource, 0)

	apps, err := listQBusinessApplications(svc)
	if err != nil {
		return nil, err
	}

	for _, appID := range apps {
		idxParams := &qbusiness.ListIndicesInput{
			ApplicationId: appID,
			MaxResults:    aws.Int64(100),
		}
		for {
			idxResp, err := svc.ListIndices(idxParams)
			if err != nil {
				return nil, err
			}
			for _, idx := range idxResp.Indices {
				dsParams := &qbusiness.ListDataSourcesInput{
					ApplicationId: appID,
					IndexId:       idx.IndexId,
					MaxResults:    aws.Int64(10),
				}
				for {
					dsResp, err := svc.ListDataSources(dsParams)
					if err != nil {
						return nil, err
					}
					for _, ds := range dsResp.DataSources {
						resources = append(resources, &QBusinessDataSource{
							svc:           svc,
							ApplicationID: appID,
							IndexID:       idx.IndexId,
							ID:            ds.DataSourceId,
							Name:          ds.DisplayName,
							Status:        ds.Status,
						})
					}
					if dsResp.NextToken == nil {
						break
					}
					dsParams.NextToken = dsResp.NextToken
				}
			}
			if idxResp.NextToken == nil {
				break
			}
			idxParams.NextToken = idxResp.NextToken
		}
	}

	return resources, nil
}

type QBusinessDataSource struct {
	svc           *qbusiness.QBusiness
	ApplicationID *string
	IndexID       *string
	ID            *string
	Name          *string
	Status        *string
}

func (r *QBusinessDataSource) Remove(_ context.Context) error {
	_, err := r.svc.DeleteDataSource(&qbusiness.DeleteDataSourceInput{
		ApplicationId: r.ApplicationID,
		IndexId:       r.IndexID,
		DataSourceId:  r.ID,
	})
	return err
}

func (r *QBusinessDataSource) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *QBusinessDataSource) String() string {
	return aws.StringValue(r.ID)
}
