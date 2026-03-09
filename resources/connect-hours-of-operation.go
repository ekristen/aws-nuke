package resources

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/connect"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const ConnectHoursOfOperationResource = "ConnectHoursOfOperation"

func init() {
	registry.Register(&registry.Registration{
		Name:     ConnectHoursOfOperationResource,
		Scope:    nuke.Account,
		Resource: &ConnectHoursOfOperation{},
		Lister:   &ConnectHoursOfOperationLister{},
	})
}

type ConnectHoursOfOperationLister struct{}

func (l *ConnectHoursOfOperationLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := connect.NewFromConfig(*opts.Config)
	var resources []resource.Resource

	instances, err := listConnectInstances(ctx, svc)
	if err != nil {
		return nil, err
	}

	for _, instance := range instances {
		paginator := connect.NewListHoursOfOperationsPaginator(svc, &connect.ListHoursOfOperationsInput{
			InstanceId: instance.Id,
		})

		for paginator.HasMorePages() {
			resp, err := paginator.NextPage(ctx)
			if err != nil {
				return nil, err
			}

			for _, hours := range resp.HoursOfOperationSummaryList {
				var tags map[string]string
				if hours.Arn != nil {
					tagsResp, err := svc.ListTagsForResource(ctx, &connect.ListTagsForResourceInput{
						ResourceArn: hours.Arn,
					})
					if err != nil {
						opts.Logger.Warnf("unable to fetch tags for connect hours of operation: %s", *hours.Arn)
					} else {
						tags = tagsResp.Tags
					}
				}

				resources = append(resources, &ConnectHoursOfOperation{
					svc:                svc,
					InstanceID:         instance.Id,
					HoursOfOperationID: hours.Id,
					Name:               hours.Name,
					ARN:                hours.Arn,
					Tags:               tags,
				})
			}
		}
	}

	return resources, nil
}

type ConnectHoursOfOperation struct {
	svc                *connect.Client
	InstanceID         *string
	HoursOfOperationID *string
	Name               *string
	ARN                *string
	Tags               map[string]string
}

func (r *ConnectHoursOfOperation) Filter() error {
	if r.Name != nil && *r.Name == "Basic Hours" {
		return fmt.Errorf("cannot delete default hours of operation")
	}
	return nil
}

func (r *ConnectHoursOfOperation) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteHoursOfOperation(ctx, &connect.DeleteHoursOfOperationInput{
		InstanceId:         r.InstanceID,
		HoursOfOperationId: r.HoursOfOperationID,
	})
	return err
}

func (r *ConnectHoursOfOperation) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *ConnectHoursOfOperation) String() string {
	return *r.Name
}
