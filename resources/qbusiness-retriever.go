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

const QBusinessRetrieverResource = "QBusinessRetriever"

func init() {
	registry.Register(&registry.Registration{
		Name:     QBusinessRetrieverResource,
		Scope:    nuke.Account,
		Resource: &QBusinessRetriever{},
		Lister:   &QBusinessRetrieverLister{},
	})
}

type QBusinessRetrieverLister struct{}

func (l *QBusinessRetrieverLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := qbusiness.NewFromConfig(*opts.Config)
	var resources []resource.Resource

	appIDs, err := listQBusinessApplicationIDs(ctx, svc)
	if err != nil {
		return nil, err
	}

	for _, appID := range appIDs {
		paginator := qbusiness.NewListRetrieversPaginator(svc, &qbusiness.ListRetrieversInput{
			ApplicationId: appID,
			MaxResults:    aws.Int32(50),
		})
		for paginator.HasMorePages() {
			resp, err := paginator.NextPage(ctx)
			if err != nil {
				return nil, err
			}
			for _, ret := range resp.Retrievers {
				resources = append(resources, &QBusinessRetriever{
					svc:           svc,
					ApplicationID: appID,
					ID:            ret.RetrieverId,
					Name:          ret.DisplayName,
					Status:        aws.String(string(ret.Status)),
				})
			}
		}
	}

	return resources, nil
}

type QBusinessRetriever struct {
	svc           *qbusiness.Client
	ApplicationID *string
	ID            *string
	Name          *string
	Status        *string
}

func (r *QBusinessRetriever) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteRetriever(ctx, &qbusiness.DeleteRetrieverInput{
		ApplicationId: r.ApplicationID,
		RetrieverId:   r.ID,
	})
	return err
}

func (r *QBusinessRetriever) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *QBusinessRetriever) String() string {
	return aws.ToString(r.ID)
}
