package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iotsitewise"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const IoTSiteWiseGatewayResource = "IoTSiteWiseGateway"

func init() {
	registry.Register(&registry.Registration{
		Name:   IoTSiteWiseGatewayResource,
		Scope:  nuke.Account,
		Lister: &IoTSiteWiseGatewayLister{},
	})
}

type IoTSiteWiseGatewayLister struct{}

func (l *IoTSiteWiseGatewayLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := iotsitewise.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &iotsitewise.ListGatewaysInput{
		MaxResults: aws.Int64(25),
	}

	for {
		resp, err := svc.ListGateways(params)
		if err != nil {
			return nil, err
		}
		for _, item := range resp.GatewaySummaries {
			resources = append(resources, &IoTSiteWiseGateway{
				svc:  svc,
				ID:   item.GatewayId,
				Name: item.GatewayName,
			})
		}

		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return resources, nil
}

type IoTSiteWiseGateway struct {
	svc  *iotsitewise.IoTSiteWise
	ID   *string
	Name *string
}

func (r *IoTSiteWiseGateway) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *IoTSiteWiseGateway) Remove(_ context.Context) error {
	_, err := r.svc.DeleteGateway(&iotsitewise.DeleteGatewayInput{
		GatewayId: r.ID,
	})

	return err
}

func (r *IoTSiteWiseGateway) String() string {
	return *r.ID
}
