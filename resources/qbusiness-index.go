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

func (l *QBusinessIndexLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := qbusiness.NewFromConfig(*opts.Config)
	var resources []resource.Resource

	appIDs, err := listQBusinessApplicationIDs(ctx, svc)
	if err != nil {
		return nil, err
	}

	for _, appID := range appIDs {
		paginator := qbusiness.NewListIndicesPaginator(svc, &qbusiness.ListIndicesInput{
			ApplicationId: appID,
			MaxResults:    aws.Int32(100),
		})
		for paginator.HasMorePages() {
			resp, err := paginator.NextPage(ctx)
			if err != nil {
				return nil, err
			}
			for _, idx := range resp.Indices {
				resources = append(resources, &QBusinessIndex{
					svc:           svc,
					ApplicationID: appID,
					ID:            idx.IndexId,
					Name:          idx.DisplayName,
					Status:        aws.String(string(idx.Status)),
				})
			}
		}
	}

	return resources, nil
}

type QBusinessIndex struct {
	svc           *qbusiness.Client
	ApplicationID *string
	ID            *string
	Name          *string
	Status        *string
}

func (r *QBusinessIndex) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteIndex(ctx, &qbusiness.DeleteIndexInput{
		ApplicationId: r.ApplicationID,
		IndexId:       r.ID,
	})
	return err
}

func (r *QBusinessIndex) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *QBusinessIndex) String() string {
	return aws.ToString(r.ID)
}
