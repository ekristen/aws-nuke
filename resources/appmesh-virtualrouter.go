package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/appmesh"

	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

type AppMeshVirtualRouter struct {
	svc               *appmesh.AppMesh
	meshName          *string
	virtualRouterName *string
}

const AppMeshVirtualRouterResource = "AppMeshVirtualRouter"

func init() {
	resource.Register(&resource.Registration{
		Name:   AppMeshVirtualRouterResource,
		Scope:  nuke.Account,
		Lister: &AppMeshVirtualRouterLister{},
	})
}

type AppMeshVirtualRouterLister struct{}

func (l *AppMeshVirtualRouterLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := appmesh.New(opts.Session)
	resources := make([]resource.Resource, 0)

	// Get Meshes
	var meshNames []*string
	err := svc.ListMeshesPages(
		&appmesh.ListMeshesInput{},
		func(page *appmesh.ListMeshesOutput, lastPage bool) bool {
			for _, mesh := range page.Meshes {
				meshNames = append(meshNames, mesh.MeshName)
			}
			return true
		},
	)
	if err != nil {
		return nil, err
	}

	// List VirtualRouters per Mesh
	var vrs []*appmesh.VirtualRouterRef
	for _, meshName := range meshNames {
		err = svc.ListVirtualRoutersPages(
			&appmesh.ListVirtualRoutersInput{
				MeshName: meshName,
			},
			func(page *appmesh.ListVirtualRoutersOutput, lastPage bool) bool {
				for _, vr := range page.VirtualRouters {
					vrs = append(vrs, vr)
				}
				return lastPage
			},
		)
		if err != nil {
			return nil, err
		}
	}

	// Create the resources
	for _, vr := range vrs {
		resources = append(resources, &AppMeshVirtualRouter{
			svc:               svc,
			meshName:          vr.MeshName,
			virtualRouterName: vr.VirtualRouterName,
		})
	}

	return resources, nil
}

func (f *AppMeshVirtualRouter) Remove(_ context.Context) error {
	_, err := f.svc.DeleteVirtualRouter(&appmesh.DeleteVirtualRouterInput{
		MeshName:          f.meshName,
		VirtualRouterName: f.virtualRouterName,
	})

	return err
}

func (f *AppMeshVirtualRouter) Properties() types.Properties {
	properties := types.NewProperties()
	properties.
		Set("MeshName", f.meshName).
		Set("Name", f.virtualRouterName)

	return properties
}
