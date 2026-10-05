package resources

import (
	"context"

	dmsv2 "github.com/aws/aws-sdk-go-v2/service/databasemigrationservice"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const DatabaseMigrationServiceInstanceProfileResource = "DatabaseMigrationServiceInstanceProfile"

func init() {
	registry.Register(&registry.Registration{
		Name:     DatabaseMigrationServiceInstanceProfileResource,
		Scope:    nuke.Account,
		Resource: &DatabaseMigrationServiceInstanceProfile{},
		Lister:   &DatabaseMigrationServiceInstanceProfileLister{},
		DependsOn: []string{
			DatabaseMigrationServiceMigrationProjectResource,
		},
	})
}

type DatabaseMigrationServiceInstanceProfileLister struct{}

func (l *DatabaseMigrationServiceInstanceProfileLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := dmsv2.NewFromConfig(*opts.Config)
	var resources []resource.Resource

	paginator := dmsv2.NewDescribeInstanceProfilesPaginator(svc, &dmsv2.DescribeInstanceProfilesInput{})

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, instanceProfile := range output.InstanceProfiles {
			resources = append(resources, &DatabaseMigrationServiceInstanceProfile{
				svc:  svc,
				ARN:  instanceProfile.InstanceProfileArn,
				Name: instanceProfile.InstanceProfileName,
			})
		}
	}

	return resources, nil
}

type DatabaseMigrationServiceInstanceProfile struct {
	ARN  *string
	Name *string

	svc *dmsv2.Client
}

func (r *DatabaseMigrationServiceInstanceProfile) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteInstanceProfile(ctx, &dmsv2.DeleteInstanceProfileInput{
		InstanceProfileIdentifier: r.ARN,
	})

	return err
}

func (r *DatabaseMigrationServiceInstanceProfile) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *DatabaseMigrationServiceInstanceProfile) String() string {
	return *r.ARN
}
