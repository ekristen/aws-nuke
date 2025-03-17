package resources

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigateway"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const APIGatewayDomainNameResource = "APIGatewayDomainName"

func init() {
	registry.Register(&registry.Registration{
		Name:     APIGatewayDomainNameResource,
		Scope:    nuke.Account,
		Resource: &APIGatewayDomainName{},
		Lister:   &APIGatewayDomainNameLister{},
	})
}

type APIGatewayDomainNameLister struct{}

func (l *APIGatewayDomainNameLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := apigateway.NewFromConfig(*opts.Config)
	var resources []resource.Resource

	params := &apigateway.GetDomainNamesInput{
		Limit: aws.Int32(100),
	}

	for {
		output, err := svc.GetDomainNames(ctx, params)
		if err != nil {
			return nil, err
		}

		for i := range output.Items {
			item := &output.Items[i]
			
			// Get tags for the domain
			tagsOutput, err := svc.GetTags(ctx, &apigateway.GetTagsInput{
				ResourceArn: aws.String("arn:aws:apigateway:" + opts.Config.Region + "::/domainnames/" + *item.DomainName),
			})
			if err != nil {
				return nil, err
			}

			tags := make(map[string]string)
			for key, value := range tagsOutput.Tags {
				tags[key] = value
			}

			resources = append(resources, &APIGatewayDomainName{
				svc:          svc,
				DomainName:   item.DomainName,
				DomainNameID: item.DomainNameId,
				Tags:         tags,
			})
		}

		if output.Position == nil {
			break
		}

		params.Position = output.Position
	}

	return resources, nil
}

type APIGatewayDomainName struct {
	svc          *apigateway.Client
	DomainName   *string
	DomainNameID *string
	Tags         map[string]string
}

func (f *APIGatewayDomainName) Remove(ctx context.Context) error {
	_, err := f.svc.DeleteDomainName(ctx, &apigateway.DeleteDomainNameInput{
		DomainName:   f.DomainName,
		DomainNameId: f.DomainNameID,
	})

	return err
}

func (f *APIGatewayDomainName) Properties() types.Properties {
	properties := types.NewProperties()
	
	// Add all tags with "tag:" prefix
	for key, value := range f.Tags {
		properties.Set("tag:"+key, value)
	}
	
	// Add other properties
	properties.Set("DomainName", f.DomainName)
	properties.Set("DomainNameID", f.DomainNameID)
	
	return properties
}

func (f *APIGatewayDomainName) String() string {
	return *f.DomainName
}