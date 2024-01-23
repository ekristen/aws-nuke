package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"

	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const SageMakerDomainResource = "SageMakerDomain"

func init() {
	resource.Register(&resource.Registration{
		Name:   SageMakerDomainResource,
		Scope:  nuke.Account,
		Lister: &SageMakerDomainLister{},
	})
}

type SageMakerDomainLister struct{}

func (l *SageMakerDomainLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := sagemaker.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &sagemaker.ListDomainsInput{
		MaxResults: aws.Int64(30),
	}

	for {
		resp, err := svc.ListDomains(params)
		if err != nil {
			return nil, err
		}

		for _, domain := range resp.Domains {
			resources = append(resources, &SageMakerDomain{
				svc:      svc,
				domainID: domain.DomainId,
			})
		}

		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return resources, nil
}

type SageMakerDomain struct {
	svc      *sagemaker.SageMaker
	domainID *string
}

func (f *SageMakerDomain) Remove(_ context.Context) error {
	_, err := f.svc.DeleteDomain(&sagemaker.DeleteDomainInput{
		DomainId:        f.domainID,
		RetentionPolicy: &sagemaker.RetentionPolicy{HomeEfsFileSystem: aws.String(sagemaker.RetentionTypeDelete)},
	})

	return err
}

func (f *SageMakerDomain) String() string {
	return *f.domainID
}

func (f *SageMakerDomain) Properties() types.Properties {
	properties := types.NewProperties()
	properties.
		Set("DomainID", f.domainID)
	return properties
}
