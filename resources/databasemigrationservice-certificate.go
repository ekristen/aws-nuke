package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/databasemigrationservice"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const DatabaseMigrationServiceCertificateResource = "DatabaseMigrationServiceCertificate"

func init() {
	registry.Register(&registry.Registration{
		Name:   DatabaseMigrationServiceCertificateResource,
		Scope:  nuke.Account,
		Lister: &DatabaseMigrationServiceCertificateLister{},
	})
}

type DatabaseMigrationServiceCertificateLister struct{}

func (l *DatabaseMigrationServiceCertificateLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := databasemigrationservice.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &databasemigrationservice.DescribeCertificatesInput{
		MaxRecords: aws.Int64(100),
	}

	for {
		output, err := svc.DescribeCertificates(params)
		if err != nil {
			return nil, err
		}

		for _, certificate := range output.Certificates {
			resources = append(resources, &DatabaseMigrationServiceCertificate{
				svc: svc,
				ARN: certificate.CertificateArn,
			})
		}

		if output.Marker == nil {
			break
		}

		params.Marker = output.Marker
	}

	return resources, nil
}

type DatabaseMigrationServiceCertificate struct {
	svc *databasemigrationservice.DatabaseMigrationService
	ARN *string
}

func (r *DatabaseMigrationServiceCertificate) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *DatabaseMigrationServiceCertificate) Remove(_ context.Context) error {
	_, err := r.svc.DeleteCertificate(&databasemigrationservice.DeleteCertificateInput{
		CertificateArn: r.ARN,
	})

	return err
}

func (r *DatabaseMigrationServiceCertificate) String() string {
	return *r.ARN
}
