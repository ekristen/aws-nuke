package resources

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/timestreaminfluxdb"
	influxdbtypes "github.com/aws/aws-sdk-go-v2/service/timestreaminfluxdb/types"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const TimestreamInfluxDBDbInstanceResource = "TimestreamInfluxDBDbInstance"

func init() {
	registry.Register(&registry.Registration{
		Name:     TimestreamInfluxDBDbInstanceResource,
		Scope:    nuke.Account,
		Resource: &TimestreamInfluxDBDbInstance{},
		Lister:   &TimestreamInfluxDBDbInstanceLister{},
	})
}

type TimestreamInfluxDBDbInstanceLister struct {
	svc TimestreamInfluxDBAPI
}

func (l *TimestreamInfluxDBDbInstanceLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	var resources []resource.Resource

	if l.svc == nil {
		opts := o.(*nuke.ListerOpts)
		l.svc = timestreaminfluxdb.NewFromConfig(*opts.Config)
	}

	params := &timestreaminfluxdb.ListDbInstancesInput{}
	for {
		resp, err := l.svc.ListDbInstances(ctx, params)
		if err != nil {
			return nil, err
		}

		for _, item := range resp.Items {
			var tags map[string]string
			tagsResp, err := l.svc.ListTagsForResource(ctx, &timestreaminfluxdb.ListTagsForResourceInput{
				ResourceArn: item.Arn,
			})
			if err == nil {
				tags = tagsResp.Tags
			}

			resources = append(resources, &TimestreamInfluxDBDbInstance{
				svc:    l.svc,
				ID:     item.Id,
				Name:   item.Name,
				Arn:    item.Arn,
				Status: string(item.Status),
				Tags:   tags,
			})
		}

		if resp.NextToken == nil {
			break
		}
		params.NextToken = resp.NextToken
	}

	return resources, nil
}

type TimestreamInfluxDBDbInstance struct {
	svc    TimestreamInfluxDBAPI
	ID     *string
	Name   *string
	Arn    *string
	Status string
	Tags   map[string]string
}

func (r *TimestreamInfluxDBDbInstance) Filter() error {
	switch influxdbtypes.Status(r.Status) {
	case influxdbtypes.StatusDeleted, influxdbtypes.StatusDeleting:
		return fmt.Errorf("already %s", r.Status)
	}
	return nil
}

func (r *TimestreamInfluxDBDbInstance) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteDbInstance(ctx, &timestreaminfluxdb.DeleteDbInstanceInput{
		Identifier: r.ID,
	})
	return err
}

func (r *TimestreamInfluxDBDbInstance) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *TimestreamInfluxDBDbInstance) String() string {
	return *r.Name
}
