package resources

import (
	"context"

	"github.com/gotidy/ptr"

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
		Limit: ptr.Int32(100),
	}

	for {
		output, err := svc.GetDomainNames(ctx, params)
		if err != nil {
			return nil, err
		}

		for i := range output.Items {
			item := &output.Items[i]

			var tags map[string]string

			// Get tags for the domain
			tagsOutput, err := svc.GetTags(ctx, &apigateway.GetTagsInput{
				ResourceArn: item.DomainNameArn,
			})
			if err != nil {
				opts.Logger.WithError(err).Error("failed to get tags for domain")
			} else if tagsOutput.Tags != nil {
				tags = tagsOutput.Tags
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

func (r *APIGatewayDomainName) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteDomainName(ctx, &apigateway.DeleteDomainNameInput{
		DomainName:   r.DomainName,
		DomainNameId: r.DomainNameID,
	})

	return err
}

func (r *APIGatewayDomainName) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *APIGatewayDomainName) String() string {
	return *r.DomainName
}
