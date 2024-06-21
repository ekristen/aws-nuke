package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/rolesanywhere"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

type IAMRolesAnywhereTrustAnchor struct {
	svc           *rolesanywhere.RolesAnywhere
	TrustAnchorID string
}

const IAMRolesAnywhereTrustAnchorResource = "IAMRolesAnywhereTrustAnchor"

func init() {
	registry.Register(&registry.Registration{
		Name:   IAMRolesAnywhereTrustAnchorResource,
		Scope:  nuke.Account,
		Lister: &IAMRolesAnywhereTrustAnchorLister{},
	})
}

type IAMRolesAnywhereTrustAnchorLister struct{}

func (l *IAMRolesAnywhereTrustAnchorLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := rolesanywhere.New(opts.Session)

	params := &rolesanywhere.ListTrustAnchorsInput{}
	resources := make([]resource.Resource, 0)

	for {
		resp, err := svc.ListTrustAnchors(params)
		if err != nil {
			return nil, err
		}
		for _, trustAnchor := range resp.TrustAnchors {
			resources = append(resources, &IAMRolesAnywhereTrustAnchor{
				svc:           svc,
				TrustAnchorID: *trustAnchor.TrustAnchorId,
			})
		}

		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return resources, nil
}

func (e *IAMRolesAnywhereTrustAnchor) Remove(_ context.Context) error {
	_, err := e.svc.DeleteTrustAnchor(&rolesanywhere.DeleteTrustAnchorInput{
		TrustAnchorId: &e.TrustAnchorID,
	})
	if err != nil {
		return err
	}

	return nil
}

func (e *IAMRolesAnywhereTrustAnchor) String() string {
	return e.TrustAnchorID
}

func (e *IAMRolesAnywhereTrustAnchor) Properties() types.Properties {
	return types.NewProperties().
		Set("TrustAnchorId", e.TrustAnchorID)
}
