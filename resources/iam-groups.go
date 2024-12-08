package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const IAMGroupResource = "IAMGroup"

func init() {
	registry.Register(&registry.Registration{
		Name:     IAMGroupResource,
		Scope:    nuke.Account,
		Resource: &IAMGroup{},
		Lister:   &IAMGroupLister{},
		DependsOn: []string{
			IAMUserGroupAttachmentResource,
			IAMGroupPolicyResource,
		},
		DeprecatedAliases: []string{
			"IamGroup",
		},
	})
}

type IAMGroup struct {
	svc  iamiface.IAMAPI
	id   string
	name string
	path string
}

func (e *IAMGroup) Remove(_ context.Context) error {
	_, err := e.svc.DeleteGroup(&iam.DeleteGroupInput{
		GroupName: &e.name,
	})
	if err != nil {
		return err
	}

	return nil
}

func (e *IAMGroup) String() string {
	return e.name
}

func (e *IAMGroup) Properties() types.Properties {
	return types.NewProperties().
		Set("Name", e.name).
		Set("Path", e.path).
		Set("ID", e.id)
}

// --------------

type IAMGroupLister struct{}

func (l *IAMGroupLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := iam.New(opts.Session)
	var resources []resource.Resource

	if err := svc.ListGroupsPages(nil, func(page *iam.ListGroupsOutput, lastPage bool) bool {
		for _, out := range page.Groups {
			resources = append(resources, &IAMGroup{
				svc:  svc,
				id:   *out.GroupId,
				name: *out.GroupName,
				path: *out.Path,
			})
		}
		return true
	}); err != nil {
		return nil, err
	}

	return resources, nil
}
