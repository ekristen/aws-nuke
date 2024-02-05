package resources

import (
	"context"
	"errors"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/gotidy/ptr"

	"github.com/aws/aws-sdk-go/service/managedblockchain"

	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const ManagedBlockchainMemberResource = "ManagedBlockchainMember"

func init() {
	resource.Register(&resource.Registration{
		Name:   ManagedBlockchainMemberResource,
		Scope:  nuke.Account,
		Lister: &ManagedBlockchainMemberLister{},
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
				networkId: n.Id,
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
	networkId *string
	name      *string
	member    *managedblockchain.MemberSummary
}

func (r *ManagedBlockchainMember) Remove(_ context.Context) error {
	_, err := r.svc.DeleteMember(&managedblockchain.DeleteMemberInput{
		NetworkId: r.networkId,
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
