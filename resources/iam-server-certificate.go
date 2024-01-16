package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"

	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const IAMServerCertificateResource = "IAMServerCertificate"

func init() {
	resource.Register(resource.Registration{
		Name:   IAMServerCertificateResource,
		Scope:  nuke.Account,
		Lister: &IAMServerCertificateLister{},
	})
}

type IAMServerCertificateLister struct{}

func (l *IAMServerCertificateLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := iam.New(opts.Session)

	resp, err := svc.ListServerCertificates(nil)
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, 0)
	for _, meta := range resp.ServerCertificateMetadataList {
		resources = append(resources, &IAMServerCertificate{
			svc:  svc,
			name: *meta.ServerCertificateName,
		})
	}

	return resources, nil
}

type IAMServerCertificate struct {
	svc  iamiface.IAMAPI
	name string
}

func (e *IAMServerCertificate) Remove(_ context.Context) error {
	_, err := e.svc.DeleteServerCertificate(&iam.DeleteServerCertificateInput{
		ServerCertificateName: &e.name,
	})
	if err != nil {
		return err
	}

	return nil
}

func (e *IAMServerCertificate) String() string {
	return e.name
}
