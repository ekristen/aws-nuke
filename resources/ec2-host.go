package resources

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const EC2HostResource = "EC2Host"

func init() {
	registry.Register(&registry.Registration{
		Name:     EC2HostResource,
		Scope:    nuke.Account,
		Resource: &EC2Host{},
		Lister:   &EC2HostLister{},
	})
}

type EC2HostLister struct{}

func (l *EC2HostLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := ec2.NewFromConfig(*opts.Config)
	params := &ec2.DescribeHostsInput{}
	paginator := ec2.NewDescribeHostsPaginator(svc, params)
	resources := make([]resource.Resource, 0)
	for paginator.HasMorePages() {
		resp, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for idx := range resp.Hosts {
			resources = append(resources, &EC2Host{
				svc:  svc,
				host: &resp.Hosts[idx],
			})
		}
	}

	return resources, nil
}

type EC2Host struct {
	svc  EC2HostAPI
	host *ec2types.Host
}

func (i *EC2Host) Filter() error {
	if i.host.State == ec2types.AllocationStateReleased {
		return fmt.Errorf("already released")
	}
	return nil
}

func (i *EC2Host) Remove(ctx context.Context) error {
	params := &ec2.ReleaseHostsInput{
		HostIds: []string{aws.ToString(i.host.HostId)},
	}

	output, err := i.svc.ReleaseHosts(ctx, params)
	if err != nil {
		return err
	}

	var releaseErrors []error
	for _, item := range output.Unsuccessful {
		code := "unknown error"
		message := "no error message returned"
		if item.Error != nil {
			code = aws.ToString(item.Error.Code)
			message = aws.ToString(item.Error.Message)
		}

		releaseErrors = append(releaseErrors, fmt.Errorf(
			"failed to release EC2 host %q: %s: %s",
			aws.ToString(item.ResourceId), code, message,
		))
	}

	return errors.Join(releaseErrors...)
}

func (i *EC2Host) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("Identifier", i.host.HostId)
	properties.Set("HostInstanceFamily", i.host.HostProperties.InstanceFamily)
	properties.Set("HostCores", i.host.HostProperties.Cores)
	properties.Set("HostState", i.host.State)
	properties.Set("AllocationTime", i.host.AllocationTime.Format(time.RFC3339))

	for _, tagValue := range i.host.Tags {
		properties.SetTag(tagValue.Key, tagValue.Value)
	}

	return properties
}

func (i *EC2Host) String() string {
	return *i.host.HostId
}
