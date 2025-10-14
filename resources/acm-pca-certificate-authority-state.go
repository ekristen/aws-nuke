package resources

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"            //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/acmpca" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const ACMPCACertificateAuthorityStateResource = "ACMPCACertificateAuthorityState"

func init() {
	registry.Register(&registry.Registration{
		Name:     ACMPCACertificateAuthorityStateResource,
		Scope:    nuke.Account,
		Resource: &ACMPCACertificateAuthorityState{},
		Lister:   &ACMPCACertificateAuthorityStateLister{},
	})
}

type ACMPCACertificateAuthorityStateLister struct{}

func (l *ACMPCACertificateAuthorityStateLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := acmpca.New(opts.Session)

	var resources []resource.Resource
	var tags []*acmpca.Tag

	params := &acmpca.ListCertificateAuthoritiesInput{
		MaxResults: aws.Int64(100),
	}

	for {
		resp, err := svc.ListCertificateAuthorities(params)
		if err != nil {
			return nil, err
		}

		for _, certificateAuthority := range resp.CertificateAuthorities {
			tagParams := &acmpca.ListTagsInput{
				CertificateAuthorityArn: certificateAuthority.Arn,
				MaxResults:              aws.Int64(100),
			}

			for {
				tagResp, tagErr := svc.ListTags(tagParams)
				if tagErr != nil {
					return nil, tagErr
				}

				tags = append(tags, tagResp.Tags...)

				if tagResp.NextToken == nil {
					break
				}
				tagParams.NextToken = tagResp.NextToken
			}

			resources = append(resources, &ACMPCACertificateAuthorityState{
				svc:    svc,
				ARN:    certificateAuthority.Arn,
				Status: certificateAuthority.Status,
				Tags:   tags,
			})
		}
		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}
	return resources, nil
}

type ACMPCACertificateAuthorityState struct {
	svc    *acmpca.ACMPCA
	ARN    *string       `description:"The Amazon Resource Name (ARN) that was assigned to the CA when it was created."`
	Status *string       `description:"The status of the CA, indicating whether it is active, creating, pending_certificate, disabled, or deleted."` //nolint:lll
	Tags   []*acmpca.Tag `description:"Tags associated with the CA."`
}

func (r *ACMPCACertificateAuthorityState) Remove(_ context.Context) error {
	_, err := r.svc.UpdateCertificateAuthority(&acmpca.UpdateCertificateAuthorityInput{
		CertificateAuthorityArn: r.ARN,
		Status:                  aws.String("DISABLED"),
	})

	return err
}

func (r *ACMPCACertificateAuthorityState) String() string {
	return *r.ARN
}

func (r *ACMPCACertificateAuthorityState) Filter() error {
	switch *r.Status {
	case "CREATING":
		return fmt.Errorf("available for deletion")
	case "PENDING_CERTIFICATE":
		return fmt.Errorf("available for deletion")
	case "DISABLED":
		return fmt.Errorf("available for deletion")
	case "DELETED":
		return fmt.Errorf("already deleted")
	default:
		return nil
	}
}

func (r *ACMPCACertificateAuthorityState) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}
