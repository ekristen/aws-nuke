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

const RamResourceShareResource = "RamResourceShare"

func init() {
	registry.Register(&registry.Registration{
		Name:     RamResourceShareResource,
		Scope:    nuke.Account,
		Resource: &RamResourceShare{},
		Lister:   &RamResourceShareLister{},
	})
}

type RamResourceShareLister struct {
	svc RamAPI
}

// List returns a list of all RAM Resource Shares owned by this account
func (l *RamResourceShareLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
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
			resources = append(resources, &RamResourceShare{
				svc:              l.svc,
				Name:             share.Name,
				OwningAccountId:  share.OwningAccountId,
				ResourceShareArn: share.ResourceShareArn,
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

// RamResourceShare is the resource type
type RamResourceShare struct {
	svc              RamAPI
	ResourceShareArn *string
	Name             *string
	OwningAccountId  *string
	Status           ramtypes.ResourceShareStatus
}

// Remove implements Resource
func (r *RamResourceShare) Remove(ctx context.Context) error {
	var notFound *ramtypes.ResourceArnNotFoundException

	// delete the resource share (doesn't delete the resource, just the share)
	_, err := r.svc.DeleteResourceShare(ctx, &ram.DeleteResourceShareInput{
		ResourceShareArn: r.ResourceShareArn,
	})

	if err != nil {
		// ignore not found error, the share is probably already deleted
		if !errors.As(err, &notFound) {
			return err
		}
	}

	return err
}

func (r *RamResourceShare) Filter() error {
	// Ignore
	if r.Status != ramtypes.ResourceShareStatusActive {
		return fmt.Errorf("RAM resource share status is %s, is not active", r.Status)
	}

	return nil
}

func (r *RamResourceShare) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}
