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

func (l *QBusinessDataSourceLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := qbusiness.NewFromConfig(*opts.Config)
	var resources []resource.Resource

	appIDs, err := listQBusinessApplicationIDs(ctx, svc)
	if err != nil {
		return nil, err
	}

	for _, appID := range appIDs {
		idxPaginator := qbusiness.NewListIndicesPaginator(svc, &qbusiness.ListIndicesInput{
			ApplicationId: appID,
			MaxResults:    aws.Int32(100),
		})
		for idxPaginator.HasMorePages() {
			idxResp, err := idxPaginator.NextPage(ctx)
			if err != nil {
				return nil, err
			}
			for _, idx := range idxResp.Indices {
				dsPaginator := qbusiness.NewListDataSourcesPaginator(svc, &qbusiness.ListDataSourcesInput{
					ApplicationId: appID,
					IndexId:       idx.IndexId,
					MaxResults:    aws.Int32(10),
				})
				for dsPaginator.HasMorePages() {
					dsResp, err := dsPaginator.NextPage(ctx)
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
							Status:        aws.String(string(ds.Status)),
						})
					}
				}
			}
		}
	}

	return resources, nil
}

type QBusinessDataSource struct {
	svc           *qbusiness.Client
	ApplicationID *string
	IndexID       *string
	ID            *string
	Name          *string
	Status        *string
}

func (r *QBusinessDataSource) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteDataSource(ctx, &qbusiness.DeleteDataSourceInput{
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
	return aws.ToString(r.ID)
}
