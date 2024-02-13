package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iot"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const IoTCACertificateResource = "IoTCACertificate"

func init() {
	registry.Register(&registry.Registration{
		Name:   IoTCACertificateResource,
		Scope:  nuke.Account,
		Lister: &IoTCACertificateLister{},
	})
}

type IoTCACertificateLister struct{}

func (l *IoTCACertificateLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := iot.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &iot.ListCACertificatesInput{}

	output, err := svc.ListCACertificates(params)
	if err != nil {
		return nil, err
	}

	for _, certificate := range output.Certificates {
		resources = append(resources, &IoTCACertificate{
			svc: svc,
			ID:  certificate.CertificateId,
		})
	}

	return resources, nil
}

type IoTCACertificate struct {
	svc *iot.IoT
	ID  *string
}

func (f *IoTCACertificate) Remove(_ context.Context) error {
	_, err := f.svc.UpdateCACertificate(&iot.UpdateCACertificateInput{
		CertificateId: f.ID,
		NewStatus:     aws.String("INACTIVE"),
	})
	if err != nil {
		return err
	}

	_, err = f.svc.DeleteCACertificate(&iot.DeleteCACertificateInput{
		CertificateId: f.ID,
	})

	return err
}

func (f *IoTCACertificate) String() string {
	return *f.ID
}
