package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudhsmv2"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const CloudHSMV2ClusterHSMResource = "CloudHSMV2ClusterHSM"

func init() {
	registry.Register(&registry.Registration{
		Name:   CloudHSMV2ClusterHSMResource,
		Scope:  nuke.Account,
		Lister: &CloudHSMV2ClusterHSMLister{},
	})
}

type CloudHSMV2ClusterHSMLister struct{}

func (l *CloudHSMV2ClusterHSMLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := cloudhsmv2.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &cloudhsmv2.DescribeClustersInput{
		MaxResults: aws.Int64(25),
	}

	for {
		resp, err := svc.DescribeClusters(params)
		if err != nil {
			return nil, err
		}

		for _, cluster := range resp.Clusters {
			for _, hsm := range cluster.Hsms {
				resources = append(resources, &CloudHSMV2ClusterHSM{
					svc:       svc,
					clusterID: hsm.ClusterId,
					hsmID:     hsm.HsmId,
				})
			}

		}

		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return resources, nil
}

type CloudHSMV2ClusterHSM struct {
	svc       *cloudhsmv2.CloudHSMV2
	clusterID *string
	hsmID     *string
}

func (f *CloudHSMV2ClusterHSM) Remove(_ context.Context) error {
	_, err := f.svc.DeleteHsm(&cloudhsmv2.DeleteHsmInput{
		ClusterId: f.clusterID,
		HsmId:     f.hsmID,
	})

	return err
}

func (f *CloudHSMV2ClusterHSM) String() string {
	return *f.hsmID
}
