package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/databasemigrationservice"

	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const DatabaseMigrationServiceCertificateResource = "DatabaseMigrationServiceCertificate"

func init() {
	resource.Register(resource.Registration{
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

func (f *DatabaseMigrationServiceCertificate) Remove(_ context.Context) error {
	_, err := f.svc.DeleteEndpoint(&databasemigrationservice.DeleteEndpointInput{
		EndpointArn: f.ARN,
	})

	return err
}

func (f *DatabaseMigrationServiceCertificate) String() string {
	return *f.ARN
}
