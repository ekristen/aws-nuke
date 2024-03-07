package resources

import (
	"context"
	"github.com/aws/aws-sdk-go/service/sagemaker/sagemakeriface"
	"github.com/gotidy/ptr"
	"github.com/sirupsen/logrus"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const SageMakerDomainResource = "SageMakerDomain"

func init() {
	registry.Register(&registry.Registration{
		Name:   SageMakerDomainResource,
		Scope:  nuke.Account,
		Lister: &SageMakerDomainLister{},
	})
}

type SageMakerDomainLister struct {
	svc sagemakeriface.SageMakerAPI
}

func (l *SageMakerDomainLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	// Note: this allows us to override svc in tests with a mock
	if l.svc == nil {
		l.svc = sagemaker.New(opts.Session)
	}

	resources := make([]resource.Resource, 0)

	params := &sagemaker.ListDomainsInput{
		MaxResults: aws.Int64(30),
	}

	for {
		resp, err := l.svc.ListDomains(params)
		if err != nil {
			return nil, err
		}

		for _, domain := range resp.Domains {
			tags := make([]*sagemaker.Tag, 0)
			tagParams := &sagemaker.ListTagsInput{
				ResourceArn: domain.DomainArn,
			}
			tagOutput, err := l.svc.ListTags(tagParams)
			if err != nil {
				logrus.WithError(err).Errorf("unable to get tags for SageMakerDomain: %s", ptr.ToString(domain.DomainId))
			}
			if tagOutput != nil {
				tags = tagOutput.Tags
			}

			resources = append(resources, &SageMakerDomain{
				svc:          l.svc,
				domainID:     domain.DomainId,
				creationTime: domain.CreationTime,
				tags:         tags,
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
	svc          sagemakeriface.SageMakerAPI
	domainID     *string
	creationTime *time.Time
	tags         []*sagemaker.Tag
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

	properties.Set("DomainID", f.domainID)
	properties.Set("CreationTime", f.creationTime.Format(time.RFC3339))

	for _, tag := range f.tags {
		properties.SetTag(tag.Key, tag.Value)
	}

	return properties
}
