package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/appmesh"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const AppMeshVirtualServiceResource = "AppMeshVirtualService"

func init() {
	registry.Register(&registry.Registration{
		Name:     AppMeshVirtualServiceResource,
		Scope:    nuke.Account,
		Resource: &AppMeshVirtualService{},
		Lister:   &AppMeshVirtualServiceLister{},
	})
}

type AppMeshVirtualServiceLister struct{}

func (l *AppMeshVirtualServiceLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
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

	// List VirtualServices per Mesh
	var vss []*appmesh.VirtualServiceRef
	for _, meshName := range meshNames {
		err = svc.ListVirtualServicesPages(
			&appmesh.ListVirtualServicesInput{
				MeshName: meshName,
			},
			func(page *appmesh.ListVirtualServicesOutput, lastPage bool) bool {
				vss = append(vss, page.VirtualServices...)
				return lastPage
			},
		)
		if err != nil {
			return nil, err
		}
	}

	// Create the resources
	for _, vs := range vss {
		resources = append(resources, &AppMeshVirtualService{
			svc:                svc,
			meshName:           vs.MeshName,
			virtualServiceName: vs.VirtualServiceName,
		})
	}

	return resources, nil
}

type AppMeshVirtualService struct {
	svc                *appmesh.AppMesh
	meshName           *string
	virtualServiceName *string
}

func (f *AppMeshVirtualService) Remove(_ context.Context) error {
	_, err := f.svc.DeleteVirtualService(&appmesh.DeleteVirtualServiceInput{
		MeshName:           f.meshName,
		VirtualServiceName: f.virtualServiceName,
	})

	return err
}

func (f *AppMeshVirtualService) Properties() types.Properties {
	properties := types.NewProperties()
	properties.
		Set("MeshName", f.meshName).
		Set("Name", f.virtualServiceName)

	return properties
}
