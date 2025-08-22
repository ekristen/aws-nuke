package resources

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const EC2VerifiedAccessEndpointResource = "EC2VerifiedAccessEndpoint"

func init() {
	registry.Register(&registry.Registration{
		Name:     EC2VerifiedAccessEndpointResource,
		Scope:    nuke.Account,
		Resource: &EC2VerifiedAccessEndpoint{},
		Lister:   &EC2VerifiedAccessEndpointLister{},
	})
}

type EC2VerifiedAccessEndpointLister struct{}

func (l *EC2VerifiedAccessEndpointLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := ec2.NewFromConfig(*opts.Config)

	params := &ec2.DescribeVerifiedAccessEndpointsInput{}
	resources := make([]resource.Resource, 0)

	for {
		resp, err := svc.DescribeVerifiedAccessEndpoints(ctx, params)
		if err != nil {
			return nil, err
		}

		for _, endpoint := range resp.VerifiedAccessEndpoints {
			resources = append(resources, &EC2VerifiedAccessEndpoint{
				svc:      svc,
				endpoint: &endpoint,
			})
		}

		if resp.NextToken == nil {
			break
		}
		params.NextToken = resp.NextToken
	}

	return resources, nil
}

type EC2VerifiedAccessEndpoint struct {
	svc      *ec2.Client
	endpoint *ec2types.VerifiedAccessEndpoint
}

func (r *EC2VerifiedAccessEndpoint) Remove(ctx context.Context) error {
	params := &ec2.DeleteVerifiedAccessEndpointInput{
		VerifiedAccessEndpointId: r.endpoint.VerifiedAccessEndpointId,
	}

	_, err := r.svc.DeleteVerifiedAccessEndpoint(ctx, params)
	return err
}

func (r *EC2VerifiedAccessEndpoint) Properties() types.Properties {
	properties := types.NewProperties()

	for _, tag := range r.endpoint.Tags {
		properties.SetTag(tag.Key, tag.Value)
	}

	properties.Set("ID", r.endpoint.VerifiedAccessEndpointId)
	properties.Set("InstanceID", r.endpoint.VerifiedAccessInstanceId)
	properties.Set("GroupID", r.endpoint.VerifiedAccessGroupId)
	properties.Set("EndpointType", r.endpoint.EndpointType)
	properties.Set("ApplicationDomain", r.endpoint.ApplicationDomain)
	properties.Set("EndpointDomain", r.endpoint.EndpointDomain)
	properties.Set("Description", r.endpoint.Description)
	properties.Set("CreationTime", r.endpoint.CreationTime)
	properties.Set("LastUpdatedTime", r.endpoint.LastUpdatedTime)
	properties.Set("Status", r.endpoint.Status)

	if r.endpoint.AttachmentType != "" {
		properties.Set("AttachmentType", r.endpoint.AttachmentType)
	}

	if r.endpoint.DomainCertificateArn != nil {
		properties.Set("DomainCertificateArn", r.endpoint.DomainCertificateArn)
	}

	return properties
}

func (r *EC2VerifiedAccessEndpoint) String() string {
	return *r.endpoint.VerifiedAccessEndpointId
}
