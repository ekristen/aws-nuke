package resources

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/lakeformation"
	lakeformationtypes "github.com/aws/aws-sdk-go-v2/service/lakeformation/types"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const LakeFormationPermissionResource = "LakeFormationPermission"

func init() {
	registry.Register(&registry.Registration{
		Name:     LakeFormationPermissionResource,
		Scope:    nuke.Account,
		Resource: &LakeFormationPermission{},
		Lister:   &LakeFormationPermissionLister{},
	})
}

type LakeFormationPermissionLister struct{}

func (l *LakeFormationPermissionLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := lakeformation.NewFromConfig(*opts.Config)
	resources := make([]resource.Resource, 0)

	paginator := lakeformation.NewListPermissionsPaginator(svc, &lakeformation.ListPermissionsInput{})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, prp := range page.PrincipalResourcePermissions {
			resources = append(resources, &LakeFormationPermission{
				svc:          svc,
				PrincipalARN: prp.Principal.DataLakePrincipalIdentifier,
				Resource:     prp.Resource,
				Permissions:  prp.Permissions,
			})
		}
	}

	return resources, nil
}

type LakeFormationPermission struct {
	svc          *lakeformation.Client
	PrincipalARN *string                         `description:"The ARN of the principal to remove permissions from"`
	Permissions  []lakeformationtypes.Permission `description:"The permissions to remove from the principal"`
	Resource     *lakeformationtypes.Resource    `description:"-"`
}

func (r *LakeFormationPermission) Remove(ctx context.Context) error {
	_, err := r.svc.RevokePermissions(ctx, &lakeformation.RevokePermissionsInput{
		Principal: &lakeformationtypes.DataLakePrincipal{
			DataLakePrincipalIdentifier: r.PrincipalARN,
		},
		Resource:    r.Resource,
		Permissions: r.Permissions,
	})

	return err
}

func (r *LakeFormationPermission) Filter() error {
	if *r.PrincipalARN == "IAM_ALLOWED_PRINCIPALS" {
		return fmt.Errorf("cannot delete default setting group permissions")
	}
	return nil
}

func (r *LakeFormationPermission) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *LakeFormationPermission) String() string {
	return *r.PrincipalARN
}
