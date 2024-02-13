package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/appmesh"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const AppMeshVirtualGatewayResource = "AppMeshVirtualGateway"

func init() {
	registry.Register(&registry.Registration{
		Name:   AppMeshVirtualGatewayResource,
		Scope:  nuke.Account,
		Lister: &AppMeshVirtualGatewayLister{},
	})
}

type AppMeshVirtualGatewayLister struct{}

func (l *AppMeshVirtualGatewayLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
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
				for _, vg := range page.VirtualGateways {
					vgs = append(vgs, vg)
				}
				return lastPage
			},
		)
		if err != nil {
			return nil, err
		}
	}

	// Create the resources
	for _, vg := range vgs {
		resources = append(resources, &AppMeshVirtualGateway{
			svc:                svc,
			meshName:           vg.MeshName,
			virtualGatewayName: vg.VirtualGatewayName,
		})
	}

	return resources, nil
}

type AppMeshVirtualGateway struct {
	svc                *appmesh.AppMesh
	meshName           *string
	virtualGatewayName *string
}

func (f *AppMeshVirtualGateway) Remove(_ context.Context) error {
	_, err := f.svc.DeleteVirtualGateway(&appmesh.DeleteVirtualGatewayInput{
		MeshName:           f.meshName,
		VirtualGatewayName: f.virtualGatewayName,
	})

	return err
}

func (f *AppMeshVirtualGateway) Properties() types.Properties {
	properties := types.NewProperties()
	properties.
		Set("MeshName", f.meshName).
		Set("Name", f.virtualGatewayName)

	return properties
}
