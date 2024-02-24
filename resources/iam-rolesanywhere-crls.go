package resources

import (
	"context"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"

	"github.com/aws/aws-sdk-go/service/rolesanywhere"
)

type IAMRolesAnywhereCRL struct {
	svc   *rolesanywhere.RolesAnywhere
	CrlID string
}

const IAMRolesAnywhereCRLResource = "IAMRolesAnywhereCRL"

func init() {
	registry.Register(&registry.Registration{
		Name:   IAMRolesAnywhereCRLResource,
		Scope:  nuke.Account,
		Lister: &IAMRolesAnywhereCRLLister{},
	})
}

type IAMRolesAnywhereCRLLister struct{}

func (l *IAMRolesAnywhereCRLLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := rolesanywhere.New(opts.Session)

	params := &rolesanywhere.ListCrlsInput{}
	resources := make([]resource.Resource, 0)

	for {
		resp, err := svc.ListCrls(params)
		if err != nil {
			return nil, err
		}
		for _, crl := range resp.Crls {
			resources = append(resources, &IAMRolesAnywhereCRL{
				svc:   svc,
				CrlID: *crl.CrlId,
			})
		}

		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return resources, nil
}

func (e *IAMRolesAnywhereCRL) Remove(_ context.Context) error {
	_, err := e.svc.DeleteCrl(&rolesanywhere.DeleteCrlInput{
		CrlId: &e.CrlID,
	})
	if err != nil {
		return err
	}

	return nil
}

func (e *IAMRolesAnywhereCRL) String() string {
	return e.CrlID
}

func (e *IAMRolesAnywhereCRL) Properties() types.Properties {
	return types.NewProperties().
		Set("CrlId", e.CrlID)
}
