package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kafka"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const MSKClusterResource = "MSKCluster"

func init() {
	registry.Register(&registry.Registration{
		Name:     MSKClusterResource,
		Scope:    nuke.Account,
		Resource: &MSKCluster{},
		Lister:   &MSKClusterLister{},
	})
}

type MSKClusterLister struct{}

func (l *MSKClusterLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := kafka.New(opts.Session)

	resp, err := svc.ListClusters(&kafka.ListClustersInput{})
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, 0)
	for _, cluster := range resp.ClusterInfoList {
		resources = append(resources, &MSKCluster{
			svc:  svc,
			arn:  *cluster.ClusterArn,
			name: *cluster.ClusterName,
			tags: cluster.Tags,
		})
	}

	return resources, nil
}

type MSKCluster struct {
	svc  *kafka.Kafka
	arn  string
	name string
	tags map[string]*string
}

func (m *MSKCluster) Remove(_ context.Context) error {
	params := &kafka.DeleteClusterInput{
		ClusterArn: &m.arn,
	}

	_, err := m.svc.DeleteCluster(params)
	if err != nil {
		return err
	}
	return nil
}

func (m *MSKCluster) String() string {
	return m.arn
}

func (m *MSKCluster) Properties() types.Properties {
	properties := types.NewProperties()
	for tagKey, tagValue := range m.tags {
		properties.SetTag(aws.String(tagKey), tagValue)
	}
	properties.Set("ARN", m.arn)
	properties.Set("Name", m.name)

	return properties
}
