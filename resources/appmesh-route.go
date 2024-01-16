package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/appmesh"

	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const AppMeshRouteResource = "AppMeshRoute"

func init() {
	resource.Register(resource.Registration{
		Name:   AppMeshRouteResource,
		Scope:  nuke.Account,
		Lister: &AppMeshRouteLister{},
	})
}

type AppMeshRouteLister struct{}

func (l *AppMeshRouteLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := appmesh.New(opts.Session)
	var resources []resource.Resource

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

	// Get VirtualRouters per mesh
	var virtualRouters []*appmesh.VirtualRouterRef
	for _, meshName := range meshNames {
		err = svc.ListVirtualRoutersPages(
			&appmesh.ListVirtualRoutersInput{
				MeshName: meshName,
			},
			func(page *appmesh.ListVirtualRoutersOutput, lastPage bool) bool {
				for _, vr := range page.VirtualRouters {
					virtualRouters = append(virtualRouters, vr)
				}
				return lastPage
			},
		)
		if err != nil {
			return nil, err
		}
	}

	// List Routes per Mesh
	var routes []*appmesh.RouteRef
	for _, vr := range virtualRouters {
		err := svc.ListRoutesPages(
			&appmesh.ListRoutesInput{
				MeshName:          vr.MeshName,
				VirtualRouterName: vr.VirtualRouterName,
			},
			func(page *appmesh.ListRoutesOutput, lastPage bool) bool {
				routes = append(routes, page.Routes...)
				return lastPage
			},
		)
		if err != nil {
			return nil, err
		}
	}

	// Create the resources
	for _, r := range routes {
		resources = append(resources, &AppMeshRoute{
			svc:               svc,
			routeName:         r.RouteName,
			meshName:          r.MeshName,
			virtualRouterName: r.VirtualRouterName,
		})
	}

	return resources, nil
}

type AppMeshRoute struct {
	svc               *appmesh.AppMesh
	routeName         *string
	meshName          *string
	virtualRouterName *string
}

func (f *AppMeshRoute) Remove(_ context.Context) error {
	_, err := f.svc.DeleteRoute(&appmesh.DeleteRouteInput{
		MeshName:          f.meshName,
		RouteName:         f.routeName,
		VirtualRouterName: f.virtualRouterName,
	})

	return err
}

func (f *AppMeshRoute) Properties() types.Properties {
	properties := types.NewProperties()
	properties.
		Set("MeshName", f.meshName).
		Set("VirtualRouterName", f.virtualRouterName).
		Set("Name", f.routeName)

	return properties
}
