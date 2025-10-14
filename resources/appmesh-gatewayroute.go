package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/appmesh" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const AppMeshGatewayRouteResource = "AppMeshGatewayRoute"

func init() {
	registry.Register(&registry.Registration{
		Name:     AppMeshGatewayRouteResource,
		Scope:    nuke.Account,
		Resource: &AppMeshGatewayRoute{},
		Lister:   &AppMeshGatewayRouteLister{},
	})
}

type AppMeshGatewayRouteLister struct{}

func (l *AppMeshGatewayRouteLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
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

	// List VirtualGateways per Mesh
	var vgs []*appmesh.VirtualGatewayRef
	for _, meshName := range meshNames {
		err = svc.ListVirtualGatewaysPages(
			&appmesh.ListVirtualGatewaysInput{
				MeshName: meshName,
			},
			func(page *appmesh.ListVirtualGatewaysOutput, lastPage bool) bool {
				vgs = append(vgs, page.VirtualGateways...)
				return lastPage
			},
		)
		if err != nil {
			return nil, err
		}
	}

	// List GatewayRoutes per VirtualGateway
	var routes []*appmesh.GatewayRouteRef
	for _, vg := range vgs {
		err := svc.ListGatewayRoutesPages(
			&appmesh.ListGatewayRoutesInput{
				MeshName:           vg.MeshName,
				VirtualGatewayName: vg.VirtualGatewayName,
			},
			func(page *appmesh.ListGatewayRoutesOutput, lastPage bool) bool {
				routes = append(routes, page.GatewayRoutes...)
				return lastPage
			},
		)
		if err != nil {
			return nil, err
		}
	}

	// Create the resources
	for _, r := range routes {
		resources = append(resources, &AppMeshGatewayRoute{
			svc:                svc,
			routeName:          r.GatewayRouteName,
			meshName:           r.MeshName,
			virtualGatewayName: r.VirtualGatewayName,
		})
	}

	return resources, nil
}

type AppMeshGatewayRoute struct {
	svc                *appmesh.AppMesh
	routeName          *string
	meshName           *string
	virtualGatewayName *string
}

func (f *AppMeshGatewayRoute) Remove(_ context.Context) error {
	_, err := f.svc.DeleteGatewayRoute(&appmesh.DeleteGatewayRouteInput{
		MeshName:           f.meshName,
		GatewayRouteName:   f.routeName,
		VirtualGatewayName: f.virtualGatewayName,
	})

	return err
}

func (f *AppMeshGatewayRoute) Properties() types.Properties {
	properties := types.NewProperties()
	properties.
		Set("MeshName", f.meshName).
		Set("VirtualGatewayName", f.virtualGatewayName).
		Set("Name", f.routeName)

	return properties
}
