package resources

import (
	"context"

	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicecatalog"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/awsutil"
	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const ServiceCatalogTagOptionPortfolioAttachmentResource = "ServiceCatalogTagOptionPortfolioAttachment"

func init() {
	registry.Register(&registry.Registration{
		Name:   ServiceCatalogTagOptionPortfolioAttachmentResource,
		Scope:  nuke.Account,
		Lister: &ServiceCatalogTagOptionPortfolioAttachmentLister{},
	})
}

type ServiceCatalogTagOptionPortfolioAttachmentLister struct{}

func (l *ServiceCatalogTagOptionPortfolioAttachmentLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := servicecatalog.New(opts.Session)
	resources := make([]resource.Resource, 0)
	var tagOptions []*servicecatalog.TagOptionDetail

	params := &servicecatalog.ListTagOptionsInput{
		PageSize: aws.Int64(20),
	}

	// list all tag options
	for {
		resp, err := svc.ListTagOptions(params)
		if err != nil {
			if awsutil.IsAWSError(err, servicecatalog.ErrCodeTagOptionNotMigratedException) {
				logrus.Info(err)
				break
			}
			return nil, err
		}

		tagOptions = append(tagOptions, resp.TagOptionDetails...)

		if resp.PageToken == nil {
			break
		}

		params.PageToken = resp.PageToken
	}

	resourceParams := &servicecatalog.ListResourcesForTagOptionInput{
		PageSize: aws.Int64(20),
	}

	for _, tagOption := range tagOptions {
		resourceParams.TagOptionId = tagOption.Id
		resp, err := svc.ListResourcesForTagOption(resourceParams)
		if err != nil {
			return nil, err
		}

		for _, resourceDetail := range resp.ResourceDetails {
			resources = append(resources, &ServiceCatalogTagOptionPortfolioAttachment{
				svc:            svc,
				tagOptionID:    tagOption.Id,
				resourceID:     resourceDetail.Id,
				resourceName:   resourceDetail.Name,
				tagOptionKey:   tagOption.Key,
				tagOptionValue: tagOption.Value,
			})
		}

		if resp.PageToken == nil {
			break
		}

		resourceParams.PageToken = resp.PageToken
	}

	return resources, nil
}

type ServiceCatalogTagOptionPortfolioAttachment struct {
	svc            *servicecatalog.ServiceCatalog
	tagOptionID    *string
	resourceID     *string
	tagOptionKey   *string
	tagOptionValue *string
	resourceName   *string
}

func (f *ServiceCatalogTagOptionPortfolioAttachment) Remove(_ context.Context) error {
	_, err := f.svc.DisassociateTagOptionFromResource(&servicecatalog.DisassociateTagOptionFromResourceInput{
		TagOptionId: f.tagOptionID,
		ResourceId:  f.resourceID,
	})

	return err
}

func (f *ServiceCatalogTagOptionPortfolioAttachment) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("TagOptionID", f.tagOptionID)
	properties.Set("TagOptionKey", f.tagOptionKey)
	properties.Set("TagOptionValue", f.tagOptionValue)
	properties.Set("ResourceID", f.resourceID)
	properties.Set("ResourceName", f.resourceName)
	return properties
}

func (f *ServiceCatalogTagOptionPortfolioAttachment) String() string {
	return fmt.Sprintf("%s -> %s", *f.tagOptionID, *f.resourceID)
}
