package resources

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/service/ec2" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/awsutil"
	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const EC2TGWAttachmentResource = "EC2TGWAttachment"

func init() {
	registry.Register(&registry.Registration{
		Name:     EC2TGWAttachmentResource,
		Scope:    nuke.Account,
		Resource: &EC2TGWAttachment{},
		Lister:   &EC2TGWAttachmentLister{},
	})
}

type EC2TGWAttachmentLister struct{}

func (l *EC2TGWAttachmentLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := ec2.New(opts.Session)

	params := &ec2.DescribeTransitGatewayAttachmentsInput{}
	resources := make([]resource.Resource, 0)
	for {
		resp, err := svc.DescribeTransitGatewayAttachments(params)
		if err != nil {
			return nil, err
		}

		for _, tgwa := range resp.TransitGatewayAttachments {
			resources = append(resources, &EC2TGWAttachment{
				svc:  svc,
				tgwa: tgwa,
			})
		}

		if resp.NextToken == nil {
			break
		}

		params = &ec2.DescribeTransitGatewayAttachmentsInput{
			NextToken: resp.NextToken,
		}
	}

	return resources, nil
}

type EC2TGWAttachment struct {
	svc  *ec2.EC2
	tgwa *ec2.TransitGatewayAttachment
}

func (e *EC2TGWAttachment) Remove(_ context.Context) error {
	if *e.tgwa.ResourceType == "VPN" {
		// This will get deleted as part of EC2VPNConnection, there is no API
		// as part of TGW to delete VPN attachments.
		return fmt.Errorf("VPN attachment")
	}
	params := &ec2.DeleteTransitGatewayVpcAttachmentInput{
		TransitGatewayAttachmentId: e.tgwa.TransitGatewayAttachmentId,
	}

	_, err := e.svc.DeleteTransitGatewayVpcAttachment(params)
	if err != nil {
		return err
	}

	return nil
}

func (e *EC2TGWAttachment) Filter() error {
	if *e.tgwa.State == awsutil.StateDeleted {
		return fmt.Errorf("already deleted")
	}

	return nil
}

func (e *EC2TGWAttachment) Properties() types.Properties {
	properties := types.NewProperties()
	for _, tagValue := range e.tgwa.Tags {
		properties.SetTag(tagValue.Key, tagValue.Value)
	}
	properties.Set("ID", e.tgwa.TransitGatewayAttachmentId)
	return properties
}

func (e *EC2TGWAttachment) String() string {
	return fmt.Sprintf("%s(%s)", *e.tgwa.TransitGatewayAttachmentId, *e.tgwa.ResourceType)
}
