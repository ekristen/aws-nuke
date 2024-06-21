package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/lambda"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const LambdaLayerResource = "LambdaLayer"

func init() {
	registry.Register(&registry.Registration{
		Name:   LambdaLayerResource,
		Scope:  nuke.Account,
		Lister: &LambdaLayerLister{},
	})
}

type LambdaLayerLister struct{}

func (l *LambdaLayerLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := lambda.New(opts.Session)

	layers := make([]*lambda.LayersListItem, 0)

	params := &lambda.ListLayersInput{}

	err := svc.ListLayersPages(params, func(page *lambda.ListLayersOutput, lastPage bool) bool {
		layers = append(layers, page.Layers...)
		return true
	})
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, 0)

	for _, layer := range layers {
		versionsParams := &lambda.ListLayerVersionsInput{
			LayerName: layer.LayerName,
		}
		err := svc.ListLayerVersionsPages(versionsParams, func(page *lambda.ListLayerVersionsOutput, lastPage bool) bool {
			for _, out := range page.LayerVersions {
				resources = append(resources, &LambdaLayer{
					svc:       svc,
					layerName: layer.LayerName,
					version:   *out.Version,
				})
			}
			return true
		})
		if err != nil {
			return nil, err
		}
	}

	return resources, nil
}

type LambdaLayer struct {
	svc       *lambda.Lambda
	layerName *string
	version   int64
}

func (l *LambdaLayer) Remove(_ context.Context) error {
	_, err := l.svc.DeleteLayerVersion(&lambda.DeleteLayerVersionInput{
		LayerName:     l.layerName,
		VersionNumber: &l.version,
	})

	return err
}

func (l *LambdaLayer) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("Name", l.layerName)
	properties.Set("Version", l.version)

	return properties
}

func (l *LambdaLayer) String() string {
	return *l.layerName
}
