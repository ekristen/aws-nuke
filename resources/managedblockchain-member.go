package resources

import (
	"context"
	"errors"

	"github.com/gotidy/ptr"

	"github.com/aws/aws-sdk-go/aws/awserr"                //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/managedblockchain" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const ManagedBlockchainMemberResource = "ManagedBlockchainMember"

func init() {
	registry.Register(&registry.Registration{
		Name:     ManagedBlockchainMemberResource,
		Scope:    nuke.Account,
		Resource: &ManagedBlockchainMember{},
		Lister:   &ManagedBlockchainMemberLister{},
	})
}

type ManagedBlockchainMemberLister struct{}

func (l *ManagedBlockchainMemberLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := managedblockchain.New(opts.Session)
	var resources []resource.Resource

	networks, err := svc.ListNetworks(&managedblockchain.ListNetworksInput{})
	if err != nil {
		return nil, err
	}

	for _, n := range networks.Networks {
		res, err := svc.ListMembers(&managedblockchain.ListMembersInput{
			NetworkId: n.Id,
		})
		if err != nil {
			return nil, err
		}

		for _, r := range res.Members {
			resources = append(resources, &ManagedBlockchainMember{
				svc:       svc,
				id:        r.Id,
				networkID: n.Id,
				name:      r.Name,
				member:    r,
			})
		}
	}

	return resources, nil
}

type ManagedBlockchainMember struct {
	svc       *managedblockchain.ManagedBlockchain
	id        *string
	networkID *string
	name      *string
	member    *managedblockchain.MemberSummary
}

func (r *ManagedBlockchainMember) Remove(_ context.Context) error {
	_, err := r.svc.DeleteMember(&managedblockchain.DeleteMemberInput{
		NetworkId: r.networkID,
		MemberId:  r.id,
	})
	if err != nil {
		var awsError awserr.Error
		if errors.As(err, &awsError) {
			if awsError.Code() == "ResourceNotFoundException" {
				return nil
			}
		}
		return err
	}

	return nil
}

func (r *ManagedBlockchainMember) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("ID", r.id)
	properties.Set("Name", r.name)
	properties.Set("CreationDate", r.member.CreationDate)
	properties.Set("IsOwned", r.member.IsOwned)
	properties.Set("Status", r.member.Status)
	return properties
}

func (r *ManagedBlockchainMember) String() string {
	return ptr.ToString(r.name)
}
