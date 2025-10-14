package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"            //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/kendra" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const KendraIndexResource = "KendraIndex"

func init() {
	registry.Register(&registry.Registration{
		Name:     KendraIndexResource,
		Scope:    nuke.Account,
		Resource: &KendraIndex{},
		Lister:   &KendraIndexLister{},
	})
}

type KendraIndexLister struct{}

func (l *KendraIndexLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := kendra.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &kendra.ListIndicesInput{
		MaxResults: aws.Int64(100),
	}

	for {
		resp, err := svc.ListIndices(params)
		if err != nil {
			return nil, err
		}
		for _, index := range resp.IndexConfigurationSummaryItems {
			resources = append(resources, &KendraIndex{
				svc:  svc,
				id:   *index.Id,
				name: *index.Name,
			})
		}

		if resp.NextToken == nil {
			break
		}
		params.NextToken = resp.NextToken
	}
	return resources, nil
}

type KendraIndex struct {
	svc  *kendra.Kendra
	name string
	id   string
}

func (i *KendraIndex) Remove(_ context.Context) error {
	_, err := i.svc.DeleteIndex(&kendra.DeleteIndexInput{
		Id: &i.id,
	})
	return err
}

func (i *KendraIndex) String() string {
	return i.id
}

func (i *KendraIndex) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("Name", i.name)

	return properties
}
