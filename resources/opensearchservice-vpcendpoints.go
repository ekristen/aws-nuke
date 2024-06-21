package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/opensearchservice"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const OSVPCEndpointResource = "OSVPCEndpoint"

func init() {
	registry.Register(&registry.Registration{
		Name:   OSVPCEndpointResource,
		Scope:  nuke.Account,
		Lister: &OSVPCEndpointLister{},
	})
}

type OSVPCEndpointLister struct{}

func (l *OSVPCEndpointLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := opensearchservice.New(opts.Session)
	resources := make([]resource.Resource, 0)
	var nextToken *string

	for {
		params := &opensearchservice.ListVpcEndpointsInput{
			NextToken: nextToken,
		}
		listResp, err := svc.ListVpcEndpoints(params)
		if err != nil {
			return nil, err
		}

		for _, vpcEndpoint := range listResp.VpcEndpointSummaryList {
			resources = append(resources, &OSVPCEndpoint{
				svc:           svc,
				vpcEndpointID: vpcEndpoint.VpcEndpointId,
			})
		}

		// Check if there are more results
		if listResp.NextToken == nil {
			break // No more results, exit the loop
		}

		// Set the nextToken for the next iteration
		nextToken = listResp.NextToken
	}

	return resources, nil
}

type OSVPCEndpoint struct {
	svc           *opensearchservice.OpenSearchService
	vpcEndpointID *string
}

func (o *OSVPCEndpoint) Remove(_ context.Context) error {
	_, err := o.svc.DeleteVpcEndpoint(&opensearchservice.DeleteVpcEndpointInput{
		VpcEndpointId: o.vpcEndpointID,
	})

	return err
}

func (o *OSVPCEndpoint) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("VpcEndpointId", o.vpcEndpointID)
	return properties
}

func (o *OSVPCEndpoint) String() string {
	return *o.vpcEndpointID
}
