package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/rolesanywhere" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

type IAMRolesAnywhereProfile struct {
	svc       *rolesanywhere.RolesAnywhere
	ProfileID string
}

const IAMRolesAnywhereProfilesResource = "IAMRolesAnywhereProfile"

func init() {
	registry.Register(&registry.Registration{
		Name:     IAMRolesAnywhereProfilesResource,
		Scope:    nuke.Account,
		Resource: &IAMRolesAnywhereProfile{},
		Lister:   &IAMRolesAnywhereProfilesLister{},
	})
}

type IAMRolesAnywhereProfilesLister struct{}

func (l *IAMRolesAnywhereProfilesLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := rolesanywhere.New(opts.Session)

	params := &rolesanywhere.ListProfilesInput{}
	resources := make([]resource.Resource, 0)

	for {
		resp, err := svc.ListProfiles(params)
		if err != nil {
			return nil, err
		}
		for _, profile := range resp.Profiles {
			resources = append(resources, &IAMRolesAnywhereProfile{
				svc:       svc,
				ProfileID: *profile.ProfileId,
			})
		}

		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return resources, nil
}

func (e *IAMRolesAnywhereProfile) Remove(_ context.Context) error {
	_, err := e.svc.DeleteProfile(&rolesanywhere.DeleteProfileInput{
		ProfileId: &e.ProfileID,
	})
	if err != nil {
		return err
	}

	return nil
}

func (e *IAMRolesAnywhereProfile) String() string {
	return e.ProfileID
}

func (e *IAMRolesAnywhereProfile) Properties() types.Properties {
	return types.NewProperties().
		Set("ProfileId", e.ProfileID)
}
