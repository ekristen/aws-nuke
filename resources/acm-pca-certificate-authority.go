package resources

import (
	"context"

	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/acmpca"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const ACMPCACertificateAuthorityResource = "ACMPCACertificateAuthority"

func init() {
	registry.Register(&registry.Registration{
		Name:                ACMPCACertificateAuthorityResource,
		Scope:               nuke.Account,
		Lister:              &ACMPCACertificateAuthorityLister{},
		AlternativeResource: "AWS::ACMPCA::CertificateAuthority",
	})
}

type ACMPCACertificateAuthorityLister struct{}

func (l *ACMPCACertificateAuthorityLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
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

			resources = append(resources, &ACMPCACertificateAuthority{
				svc:    svc,
				ARN:    certificateAuthority.Arn,
				status: certificateAuthority.Status,
				tags:   tags,
			})
		}
		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}
	return resources, nil
}

type ACMPCACertificateAuthority struct {
	svc    *acmpca.ACMPCA
	ARN    *string
	status *string
	tags   []*acmpca.Tag
}

func (r *ACMPCACertificateAuthority) Remove(_ context.Context) error {
	_, err := r.svc.DeleteCertificateAuthority(&acmpca.DeleteCertificateAuthorityInput{
		CertificateAuthorityArn: r.ARN,
	})

	return err
}

func (r *ACMPCACertificateAuthority) String() string {
	return *r.ARN
}

func (r *ACMPCACertificateAuthority) Filter() error {
	if *r.status == "DELETED" {
		return fmt.Errorf("already deleted")
	} else {
		return nil
	}
}

func (r *ACMPCACertificateAuthority) Properties() types.Properties {
	properties := types.NewProperties()
	for _, tag := range r.tags {
		properties.SetTag(tag.Key, tag.Value)
	}
	properties.
		Set("ARN", r.ARN).
		Set("Status", r.status)
	return properties
}
