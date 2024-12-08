package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/gotidy/ptr"

	"github.com/aws/aws-sdk-go/service/networkmanager"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/awsutil"
	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const NetworkManagerNetworkAttachmentResource = "NetworkManagerNetworkAttachment"

func init() {
	registry.Register(&registry.Registration{
		Name:     NetworkManagerNetworkAttachmentResource,
		Scope:    nuke.Account,
		Resource: &NetworkManagerNetworkAttachment{},
		Lister:   &NetworkManagerNetworkAttachmentLister{},
	})
}

type NetworkManagerNetworkAttachmentLister struct{}

func (l *NetworkManagerNetworkAttachmentLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := networkmanager.New(opts.Session)
	params := &networkmanager.ListAttachmentsInput{}
	resources := make([]resource.Resource, 0)

	resp, err := svc.ListAttachments(params)
	if err != nil {
		return nil, err
	}

	for _, attachment := range resp.Attachments {
		resources = append(resources, &NetworkManagerNetworkAttachment{
			svc:        svc,
			attachment: attachment,
		})
	}

	return resources, nil
}

type NetworkManagerNetworkAttachment struct {
	svc        *networkmanager.NetworkManager
	attachment *networkmanager.Attachment
}

func (n *NetworkManagerNetworkAttachment) Remove(_ context.Context) error {
	params := &networkmanager.DeleteAttachmentInput{
		AttachmentId: n.attachment.AttachmentId,
	}

	_, err := n.svc.DeleteAttachment(params)
	if err != nil {
		return err
	}

	return nil
}

func (n *NetworkManagerNetworkAttachment) Filter() error {
	if strings.EqualFold(ptr.ToString(n.attachment.State), awsutil.StateDeleted) {
		return fmt.Errorf("already deleted")
	}

	return nil
}

func (n *NetworkManagerNetworkAttachment) Properties() types.Properties {
	properties := types.NewProperties()

	properties.Set("ID", n.attachment.AttachmentId)
	properties.Set("ARN", n.attachment.ResourceArn)

	for _, tagValue := range n.attachment.Tags {
		properties.SetTag(tagValue.Key, tagValue.Value)
	}

	return properties
}

func (n *NetworkManagerNetworkAttachment) String() string {
	return *n.attachment.AttachmentId
}
