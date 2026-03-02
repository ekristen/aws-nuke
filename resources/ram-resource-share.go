package resources

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ram"
	ramtypes "github.com/aws/aws-sdk-go-v2/service/ram/types"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const RAMResourceShareResource = "RAMResourceShare"

func init() {
	registry.Register(&registry.Registration{
		Name:     RAMResourceShareResource,
		Scope:    nuke.Account,
		Resource: &RAMResourceShare{},
		Lister:   &RAMResourceShareLister{},
	})
}

type RAMResourceShareLister struct {
	svc RAMAPI
}

// List returns a list of all RAM Resource Shares owned by this account
func (l *RAMResourceShareLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	var resources []resource.Resource

	if l.svc == nil {
		opts := o.(*nuke.ListerOpts)
		svc := ram.NewFromConfig(*opts.Config)
		l.svc = svc
	}

	params := &ram.GetResourceSharesInput{
		ResourceOwner: "SELF",
	}
	for {
		resp, err := l.svc.GetResourceShares(ctx, params)

		if err != nil {
			return nil, err
		}

		for _, share := range resp.ResourceShares {
			resources = append(resources, &RAMResourceShare{
				svc:              l.svc,
				Name:             share.Name,
				OwningAccountID:  share.OwningAccountId,
				ResourceShareARN: share.ResourceShareArn,
				Status:           share.Status,
			})
		}

		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return resources, nil
}

// RAMResourceShare is the resource type
type RAMResourceShare struct {
	svc              RAMAPI
	ResourceShareARN *string
	Name             *string
	OwningAccountID  *string
	Status           ramtypes.ResourceShareStatus
}

// Remove implements Resource
func (r *RAMResourceShare) Remove(ctx context.Context) error {
	var notFound *ramtypes.ResourceArnNotFoundException

	// delete the resource share (doesn't delete the resource, just the share)
	_, err := r.svc.DeleteResourceShare(ctx, &ram.DeleteResourceShareInput{
		ResourceShareArn: r.ResourceShareARN,
	})

	if err != nil {
		if !errors.As(err, &notFound) {
			return err
		}
	}

	return err
}

func (r *RAMResourceShare) Filter() error {
	if r.Status != ramtypes.ResourceShareStatusActive {
		return fmt.Errorf("RAM resource share status is %s, is not active", r.Status)
	}

	return nil
}

func (r *RAMResourceShare) Properties() types.Properties {
	props := types.NewPropertiesFromStruct(r)
	// TODO(v4): remove backward-compat properties
	props.Set("OwningAccountId", r.OwningAccountID)
	props.Set("ResourceShareArn", r.ResourceShareARN)
	return props
}
