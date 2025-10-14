package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"             //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/appsync" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const AppSyncGraphqlAPIResource = "AppSyncGraphqlAPI"

func init() {
	registry.Register(&registry.Registration{
		Name:     AppSyncGraphqlAPIResource,
		Scope:    nuke.Account,
		Resource: &AppSyncGraphqlAPI{},
		Lister:   &AppSyncGraphqlAPILister{},
	})
}

type AppSyncGraphqlAPILister struct{}

// List - List all AWS AppSync GraphQL APIs in the account
func (l *AppSyncGraphqlAPILister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := appsync.New(opts.Session)
	var resources []resource.Resource

	params := &appsync.ListGraphqlApisInput{
		MaxResults: aws.Int64(25),
	}

	for {
		resp, err := svc.ListGraphqlApis(params)
		if err != nil {
			return nil, err
		}

		for _, graphqlAPI := range resp.GraphqlApis {
			resources = append(resources, &AppSyncGraphqlAPI{
				svc:   svc,
				apiID: graphqlAPI.ApiId,
				name:  graphqlAPI.Name,
				tags:  graphqlAPI.Tags,
			})
		}

		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return resources, nil
}

// AppSyncGraphqlAPI - An AWS AppSync GraphQL API
type AppSyncGraphqlAPI struct {
	svc   *appsync.AppSync
	apiID *string
	name  *string
	tags  map[string]*string
}

// Remove - remove an AWS AppSync GraphQL API
func (f *AppSyncGraphqlAPI) Remove(_ context.Context) error {
	_, err := f.svc.DeleteGraphqlApi(&appsync.DeleteGraphqlApiInput{
		ApiId: f.apiID,
	})
	return err
}

// Properties - Get the properties of an AWS AppSync GraphQL API
func (f *AppSyncGraphqlAPI) Properties() types.Properties {
	properties := types.NewProperties()
	for key, value := range f.tags {
		properties.SetTag(aws.String(key), value)
	}
	properties.Set("Name", f.name)
	properties.Set("APIID", f.apiID)
	return properties
}

func (f *AppSyncGraphqlAPI) String() string {
	return *f.apiID
}
