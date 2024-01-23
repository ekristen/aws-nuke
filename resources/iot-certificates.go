package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iot"

	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const IoTCertificateResource = "IoTCertificate"

func init() {
	resource.Register(&resource.Registration{
		Name:   IoTCertificateResource,
		Scope:  nuke.Account,
		Lister: &IoTCertificateLister{},
	})
}

type IoTCertificateLister struct{}

func (l *IoTCertificateLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := iot.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &iot.ListCertificatesInput{}

	for {
		output, err := svc.ListCertificates(params)
		if err != nil {
			return nil, err
		}

		for _, certificate := range output.Certificates {
			resources = append(resources, &IoTCertificate{
				svc: svc,
				ID:  certificate.CertificateId,
			})
		}
		if output.NextMarker == nil {
			break
		}

		params.Marker = output.NextMarker
	}

	return resources, nil
}

type IoTCertificate struct {
	svc *iot.IoT
	ID  *string
}

func (f *IoTCertificate) Remove(_ context.Context) error {
	_, err := f.svc.UpdateCertificate(&iot.UpdateCertificateInput{
		CertificateId: f.ID,
		NewStatus:     aws.String("INACTIVE"),
	})
	if err != nil {
		return err
	}

	_, err = f.svc.DeleteCertificate(&iot.DeleteCertificateInput{
		CertificateId: f.ID,
	})

	return err
}

func (f *IoTCertificate) String() string {
	return *f.ID
}
