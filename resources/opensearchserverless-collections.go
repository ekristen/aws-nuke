package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/opensearchserverless"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const OpenSearchServerlessCollectionResource = "OSCollection"

func init() {
	registry.Register(&registry.Registration{
		Name:   OpenSearchServerlessCollectionResource,
		Scope:  nuke.Account,
		Lister: &OSCollectionLister{},
	})
}

type OSCollectionLister struct{}

func (l *OSCollectionLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := opensearchserverless.New(opts.Session)
	resources := make([]resource.Resource, 0)
	var nextToken *string

	for {
		listResp, err := svc.ListCollections(&opensearchserverless.ListCollectionsInput{NextToken: nextToken})
		if err != nil {
			return nil, err
		}

		for _, collection := range listResp.CollectionSummaries {
			listTagsOutput, _ := svc.ListTagsForResource(&opensearchserverless.ListTagsForResourceInput{
				ResourceArn: collection.Arn,
			})
			resources = append(resources, &OSCollection{
				svc:  svc,
				id:   collection.Id,
				name: collection.Name,
				arn:  collection.Arn,
				tags: listTagsOutput.Tags,
			})
		}

		if listResp.NextToken == nil {
			break
		}

		nextToken = listResp.NextToken
	}
	return resources, nil
}

type OSCollection struct {
	svc  *opensearchserverless.OpenSearchServerless
	id   *string
	name *string
	arn  *string
	tags []*opensearchserverless.Tag
}

func (o *OSCollection) Remove(_ context.Context) error {
	_, err := o.svc.DeleteCollection(&opensearchserverless.DeleteCollectionInput{
		Id: o.id,
	})

	return err
}

func (o *OSCollection) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("ID", o.id)
	properties.Set("NAME", o.name)
	properties.Set("ARN", o.arn)
	for _, tagValue := range o.tags {
		properties.SetTag(tagValue.Key, tagValue.Value)
	}
	return properties
}

func (o *OSCollection) String() string {
	return *o.name
}
