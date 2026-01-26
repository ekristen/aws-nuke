package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/appmesh" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const AppMeshVirtualNodeResource = "AppMeshVirtualNode"

func init() {
	registry.Register(&registry.Registration{
		Name:     AppMeshVirtualNodeResource,
		Scope:    nuke.Account,
		Resource: &AppMeshVirtualNode{},
		Lister:   &AppMeshVirtualNodeLister{},
	})
}

type AppMeshVirtualNodeLister struct{}

func (l *AppMeshVirtualNodeLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
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

	// List VirtualNodes per Mesh
	var vns []*appmesh.VirtualNodeRef
	for _, meshName := range meshNames {
		err = svc.ListVirtualNodesPages(
			&appmesh.ListVirtualNodesInput{
				MeshName: meshName,
			},
			func(page *appmesh.ListVirtualNodesOutput, lastPage bool) bool {
				vns = append(vns, page.VirtualNodes...)
				return lastPage
			},
		)
		if err != nil {
			return nil, err
		}
	}

	// Create the resources
	for _, vn := range vns {
		resources = append(resources, &AppMeshVirtualNode{
			svc:             svc,
			meshName:        vn.MeshName,
			virtualNodeName: vn.VirtualNodeName,
		})
	}

	return resources, nil
}

type AppMeshVirtualNode struct {
	svc             *appmesh.AppMesh
	meshName        *string
	virtualNodeName *string
}

func (f *AppMeshVirtualNode) Remove(_ context.Context) error {
	_, err := f.svc.DeleteVirtualNode(&appmesh.DeleteVirtualNodeInput{
		MeshName:        f.meshName,
		VirtualNodeName: f.virtualNodeName,
	})

	return err
}

func (f *AppMeshVirtualNode) Properties() types.Properties {
	properties := types.NewProperties()
	properties.
		Set("MeshName", f.meshName).
		Set("Name", f.virtualNodeName)

	return properties
}

func (f *AppMeshVirtualNode) String() string {
	return *f.virtualNodeName
}
