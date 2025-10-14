package resources

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/aws"         //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/acm" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const ACMCertificateResource = "ACMCertificate"

func init() {
	registry.Register(&registry.Registration{
		Name:     ACMCertificateResource,
		Scope:    nuke.Account,
		Resource: &ACMCertificate{},
		Lister:   &ACMCertificateLister{},
	})
}

type ACMCertificateLister struct{}

func (l *ACMCertificateLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := acm.New(opts.Session)
	var resources []resource.Resource

	params := &acm.ListCertificatesInput{
		MaxItems: aws.Int64(100),
		Includes: &acm.Filters{
			KeyTypes: aws.StringSlice([]string{
				acm.KeyAlgorithmEcPrime256v1,
				acm.KeyAlgorithmEcSecp384r1,
				acm.KeyAlgorithmEcSecp521r1,
				acm.KeyAlgorithmRsa1024,
				acm.KeyAlgorithmRsa2048,
				acm.KeyAlgorithmRsa4096,
			})},
	}

	for {
		resp, err := svc.ListCertificates(params)
		if err != nil {
			return nil, err
		}

		for _, certificate := range resp.CertificateSummaryList {
			// Unfortunately the ACM API doesn't provide the certificate details when listing, so we
			// have to describe each certificate separately.
			certificateDescribe, err := svc.DescribeCertificate(&acm.DescribeCertificateInput{
				CertificateArn: certificate.CertificateArn,
			})
			if err != nil {
				return nil, err
			}

			tagParams := &acm.ListTagsForCertificateInput{
				CertificateArn: certificate.CertificateArn,
			}

			tagResp, tagErr := svc.ListTagsForCertificate(tagParams)
			if tagErr != nil {
				return nil, tagErr
			}

			resources = append(resources, &ACMCertificate{
				svc:        svc,
				ARN:        certificate.CertificateArn,
				DomainName: certificateDescribe.Certificate.DomainName,
				Status:     certificateDescribe.Certificate.Status,
				CreatedAt:  certificateDescribe.Certificate.CreatedAt,
				Tags:       tagResp.Tags,
			})
		}

		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return resources, nil
}

type ACMCertificate struct {
	svc        *acm.ACM
	ARN        *string    `description:"The ARN of the certificate"`
	DomainName *string    `description:"The domain name of the certificate"`
	Status     *string    `description:"The status of the certificate"`
	CreatedAt  *time.Time `description:"The creation time of the certificate"`
	Tags       []*acm.Tag `description:"The tags of the certificate"`
}

func (r *ACMCertificate) Remove(_ context.Context) error {
	_, err := r.svc.DeleteCertificate(&acm.DeleteCertificateInput{
		CertificateArn: r.ARN,
	})

	return err
}

func (r *ACMCertificate) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *ACMCertificate) String() string {
	return *r.ARN
}
