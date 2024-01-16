package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iot"

	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const IoTRoleAliasResource = "IoTRoleAlias"

func init() {
	resource.Register(resource.Registration{
		Name:   IoTRoleAliasResource,
		Scope:  nuke.Account,
		Lister: &IoTRoleAliasLister{},
	})
}

type IoTRoleAliasLister struct{}

func (l *IoTRoleAliasLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := iot.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &iot.ListRoleAliasesInput{
		PageSize: aws.Int64(25),
	}
	for {
		output, err := svc.ListRoleAliases(params)
		if err != nil {
			return nil, err
		}

		for _, roleAlias := range output.RoleAliases {
			resources = append(resources, &IoTRoleAlias{
				svc:       svc,
				roleAlias: roleAlias,
			})
		}
		if output.NextMarker == nil {
			break
		}

		params.Marker = output.NextMarker
	}

	return resources, nil
}

type IoTRoleAlias struct {
	svc       *iot.IoT
	roleAlias *string
}

func (f *IoTRoleAlias) Remove(_ context.Context) error {
	_, err := f.svc.DeleteRoleAlias(&iot.DeleteRoleAliasInput{
		RoleAlias: f.roleAlias,
	})

	return err
}

func (f *IoTRoleAlias) String() string {
	return *f.roleAlias
}
