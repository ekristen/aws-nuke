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

func (l *QBusinessRetrieverLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := qbusiness.New(opts.Session)
	resources := make([]resource.Resource, 0)

	apps, err := listQBusinessApplications(svc)
	if err != nil {
		return nil, err
	}

	for _, appID := range apps {
		params := &qbusiness.ListRetrieversInput{
			ApplicationId: appID,
			MaxResults:    aws.Int64(50),
		}
		for {
			resp, err := svc.ListRetrievers(params)
			if err != nil {
				return nil, err
			}
			for _, ret := range resp.Retrievers {
				resources = append(resources, &QBusinessRetriever{
					svc:           svc,
					ApplicationID: appID,
					ID:            ret.RetrieverId,
					Name:          ret.DisplayName,
					Status:        ret.Status,
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

type QBusinessRetriever struct {
	svc           *qbusiness.QBusiness
	ApplicationID *string
	ID            *string
	Name          *string
	Status        *string
}

func (r *QBusinessRetriever) Remove(_ context.Context) error {
	_, err := r.svc.DeleteRetriever(&qbusiness.DeleteRetrieverInput{
		ApplicationId: r.ApplicationID,
		RetrieverId:   r.ID,
	})
	return err
}

func (r *QBusinessRetriever) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *QBusinessRetriever) String() string {
	return aws.StringValue(r.ID)
}
