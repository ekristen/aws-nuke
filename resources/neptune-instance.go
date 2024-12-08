package resources

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/neptune"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const NeptuneInstanceResource = "NeptuneInstance"

func init() {
	registry.Register(&registry.Registration{
		Name:     NeptuneInstanceResource,
		Scope:    nuke.Account,
		Resource: &NeptuneInstance{},
		Lister:   &NeptuneInstanceLister{},
	})
}

type NeptuneInstanceLister struct{}

func (l *NeptuneInstanceLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := neptune.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &neptune.DescribeDBInstancesInput{
		MaxRecords: aws.Int64(100),
		Filters: []*neptune.Filter{
			{
				Name:   aws.String("engine"),
				Values: []*string{aws.String("neptune")},
			},
		},
	}

	for {
		output, err := svc.DescribeDBInstances(params)
		if err != nil {
			return nil, err
		}

		for _, dbInstance := range output.DBInstances {
			var dbTags []*neptune.Tag
			tags, err := svc.ListTagsForResource(&neptune.ListTagsForResourceInput{
				ResourceName: dbInstance.DBInstanceArn,
			})
			if err != nil {
				logrus.WithError(err).Warn("failed to list tags for resource")
			}
			if tags.TagList != nil {
				dbTags = tags.TagList
			}

			resources = append(resources, &NeptuneInstance{
				svc:  svc,
				ID:   dbInstance.DBInstanceIdentifier,
				Name: dbInstance.DBName,
				Tags: dbTags,
			})
		}

		if output.Marker == nil {
			break
		}

		params.Marker = output.Marker
	}

	return resources, nil
}

type NeptuneInstance struct {
	svc  *neptune.Neptune
	ID   *string
	Name *string
	Tags []*neptune.Tag
}

func (r *NeptuneInstance) Remove(_ context.Context) error {
	_, err := r.svc.DeleteDBInstance(&neptune.DeleteDBInstanceInput{
		DBInstanceIdentifier: r.ID,
		SkipFinalSnapshot:    aws.Bool(true),
	})

	return err
}

func (r *NeptuneInstance) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *NeptuneInstance) String() string {
	return *r.ID
}
