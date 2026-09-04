package resources

import (
	"context"

	dmsv2 "github.com/aws/aws-sdk-go-v2/service/databasemigrationservice"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const DatabaseMigrationServiceMigrationProjectResource = "DatabaseMigrationServiceMigrationProject"

func init() {
	registry.Register(&registry.Registration{
		Name:     DatabaseMigrationServiceMigrationProjectResource,
		Scope:    nuke.Account,
		Resource: &DatabaseMigrationServiceMigrationProject{},
		Lister:   &DatabaseMigrationServiceMigrationProjectLister{},
	})
}

type DatabaseMigrationServiceMigrationProjectLister struct{}

func (l *DatabaseMigrationServiceMigrationProjectLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := dmsv2.NewFromConfig(*opts.Config)
	var resources []resource.Resource

	paginator := dmsv2.NewDescribeMigrationProjectsPaginator(svc, &dmsv2.DescribeMigrationProjectsInput{})

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, migrationProject := range output.MigrationProjects {
			resources = append(resources, &DatabaseMigrationServiceMigrationProject{
				svc:  svc,
				ARN:  migrationProject.MigrationProjectArn,
				Name: migrationProject.MigrationProjectName,
			})
		}
	}

	return resources, nil
}

type DatabaseMigrationServiceMigrationProject struct {
	ARN  *string
	Name *string

	svc *dmsv2.Client
}

func (r *DatabaseMigrationServiceMigrationProject) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteMigrationProject(ctx, &dmsv2.DeleteMigrationProjectInput{
		MigrationProjectIdentifier: r.ARN,
	})

	return err
}

func (r *DatabaseMigrationServiceMigrationProject) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *DatabaseMigrationServiceMigrationProject) String() string {
	return *r.ARN
}
