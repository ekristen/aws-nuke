package resources

import (
	"context"

	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/service/emr"

	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const EMRClusterResource = "EMRCluster"

func init() {
	resource.Register(&resource.Registration{
		Name:   EMRClusterResource,
		Scope:  nuke.Account,
		Lister: &EMRClusterLister{},
	})
}

type EMRClusterLister struct{}

func (l *EMRClusterLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := emr.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &emr.ListClustersInput{}

	for {
		resp, err := svc.ListClusters(params)
		if err != nil {
			return nil, err
		}

		for _, cluster := range resp.Clusters {
			resources = append(resources, &EMRCluster{
				svc:     svc,
				cluster: cluster,
				state:   cluster.Status.State,
			})
		}

		if resp.Marker == nil {
			break
		}

		params.Marker = resp.Marker
	}

	return resources, nil
}

type EMRCluster struct {
	svc     *emr.EMR
	cluster *emr.ClusterSummary
	state   *string
}

func (f *EMRCluster) Filter() error {
	if strings.Contains(*f.state, "TERMINATED") {
		return fmt.Errorf("already terminated")
	}
	return nil
}

func (f *EMRCluster) Remove(_ context.Context) error {
	// Note: Call names are inconsistent in the SDK
	_, err := f.svc.TerminateJobFlows(&emr.TerminateJobFlowsInput{
		JobFlowIds: []*string{f.cluster.Id},
	})

	// Force nil return due to async callbacks blocking
	if err == nil {
		return nil
	}

	return err
}

func (f *EMRCluster) Properties() types.Properties {
	properties := types.NewProperties().
		Set("CreatedTime", f.cluster.Status.Timeline.CreationDateTime.Format(time.RFC3339))

	return properties
}

func (f *EMRCluster) String() string {
	return *f.cluster.Id
}
