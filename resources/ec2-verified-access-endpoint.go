package resources

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/gotidy/ptr"

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
				svc:                   svc,
				ID:                    endpoint.VerifiedAccessEndpointId,
				Description:           endpoint.Description,
				CreationTime:          endpoint.CreationTime,
				LastUpdatedTime:       endpoint.LastUpdatedTime,
				VerifiedAccessGroupId: endpoint.VerifiedAccessGroupId,
				ApplicationDomain:     endpoint.ApplicationDomain,
				EndpointType:          ptr.String(string(endpoint.EndpointType)),
				AttachmentType:        ptr.String(string(endpoint.AttachmentType)),
				DomainCertificateArn:  endpoint.DomainCertificateArn,
				Tags:                  endpoint.Tags,
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
	svc                   *ec2.Client
	ID                    *string
	Description           *string
	CreationTime          *string
	LastUpdatedTime       *string
	VerifiedAccessGroupId *string
	ApplicationDomain     *string
	EndpointType          *string
	AttachmentType        *string
	DomainCertificateArn  *string
	Tags                  []ec2types.Tag
}

func (r *EC2VerifiedAccessEndpoint) Remove(ctx context.Context) error {
	params := &ec2.DeleteVerifiedAccessEndpointInput{
		VerifiedAccessEndpointId: r.ID,
	}

	_, err := r.svc.DeleteVerifiedAccessEndpoint(ctx, params)
	return err
}

func (r *EC2VerifiedAccessEndpoint) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *EC2VerifiedAccessEndpoint) String() string {
	return *r.ID
}
