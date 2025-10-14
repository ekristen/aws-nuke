package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"             //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/appmesh" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const AppMeshMeshResource = "AppMeshMesh"

func init() {
	registry.Register(&registry.Registration{
		Name:     AppMeshMeshResource,
		Scope:    nuke.Account,
		Resource: &AppMeshMesh{},
		Lister:   &AppMeshMeshLister{},
	})
}

type AppMeshMeshLister struct{}

func (l *AppMeshMeshLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := appmesh.New(opts.Session)
	var resources []resource.Resource

	params := &appmesh.ListMeshesInput{
		Limit: aws.Int64(100),
	}

	for {
		output, err := svc.ListMeshes(params)
		if err != nil {
			return nil, err
		}

		for _, item := range output.Meshes {
			resources = append(resources, &AppMeshMesh{
				svc:      svc,
				meshName: item.MeshName,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type AppMeshMesh struct {
	svc      *appmesh.AppMesh
	meshName *string
}

func (f *AppMeshMesh) Remove(_ context.Context) error {
	_, err := f.svc.DeleteMesh(&appmesh.DeleteMeshInput{
		MeshName: f.meshName,
	})

	return err
}

func (f *AppMeshMesh) Properties() types.Properties {
	properties := types.NewProperties()
	properties.
		Set("MeshName", f.meshName)

	return properties
}
