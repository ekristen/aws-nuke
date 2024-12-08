package resources

import (
	"context"

	"fmt"

	"github.com/gotidy/ptr"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const IAMSigningCertificateResource = "IAMSigningCertificate"

func init() {
	registry.Register(&registry.Registration{
		Name:     IAMSigningCertificateResource,
		Scope:    nuke.Account,
		Resource: &IAMSigningCertificate{},
		Lister:   &IAMSigningCertificateLister{},
	})
}

type IAMSigningCertificateLister struct{}

func (l *IAMSigningCertificateLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := iam.New(opts.Session)
	var resources []resource.Resource

	params := &iam.ListUsersInput{
		MaxItems: aws.Int64(100),
	}

	for {
		resp, err := svc.ListUsers(params)
		if err != nil {
			return nil, err
		}

		for _, out := range resp.Users {
			resp, err := svc.ListSigningCertificates(&iam.ListSigningCertificatesInput{
				UserName: out.UserName,
			})
			if err != nil {
				return nil, err
			}

			for _, signingCert := range resp.Certificates {
				resources = append(resources, &IAMSigningCertificate{
					svc:           svc,
					certificateID: signingCert.CertificateId,
					userName:      signingCert.UserName,
					status:        signingCert.Status,
				})
			}
		}

		if resp.Marker == nil {
			break
		}

		params.Marker = resp.Marker
	}

	return resources, nil
}

type IAMSigningCertificate struct {
	svc           iamiface.IAMAPI
	certificateID *string
	userName      *string
	status        *string
}

func (i *IAMSigningCertificate) Remove(_ context.Context) error {
	_, err := i.svc.DeleteSigningCertificate(&iam.DeleteSigningCertificateInput{
		CertificateId: i.certificateID,
		UserName:      i.userName,
	})
	return err
}

func (i *IAMSigningCertificate) Properties() types.Properties {
	return types.NewProperties().
		Set("UserName", i.userName).
		Set("CertificateId", i.certificateID).
		Set("Status", i.status)
}

func (i *IAMSigningCertificate) String() string {
	return fmt.Sprintf("%s -> %s", ptr.ToString(i.userName), ptr.ToString(i.certificateID))
}
