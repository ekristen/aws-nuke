package resources

import (
	"context"

	"fmt"

	"github.com/aws/aws-sdk-go/service/efs"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const EFSMountTargetResource = "EFSMountTarget"

func init() {
	registry.Register(&registry.Registration{
		Name:   EFSMountTargetResource,
		Scope:  nuke.Account,
		Lister: &EFSMountTargetLister{},
	})
}

type EFSMountTargetLister struct{}

func (l *EFSMountTargetLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := efs.New(opts.Session)

	resp, err := svc.DescribeFileSystems(nil)
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, 0)
	for _, fs := range resp.FileSystems {
		mt, err := svc.DescribeMountTargets(&efs.DescribeMountTargetsInput{
			FileSystemId: fs.FileSystemId,
		})
		if err != nil {
			return nil, err
		}

		lto, err := svc.ListTagsForResource(&efs.ListTagsForResourceInput{ResourceId: fs.FileSystemId})
		if err != nil {
			return nil, err
		}

		for _, t := range mt.MountTargets {
			resources = append(resources, &EFSMountTarget{
				svc:    svc,
				id:     *t.MountTargetId,
				fsID:   *t.FileSystemId,
				fsTags: lto.Tags,
			})
		}
	}

	return resources, nil
}

type EFSMountTarget struct {
	svc    *efs.EFS
	id     string
	fsID   string
	fsTags []*efs.Tag
}

func (e *EFSMountTarget) Remove(_ context.Context) error {
	_, err := e.svc.DeleteMountTarget(&efs.DeleteMountTargetInput{
		MountTargetId: &e.id,
	})

	return err
}

func (e *EFSMountTarget) Properties() types.Properties {
	properties := types.NewProperties()
	for _, tagValue := range e.fsTags {
		properties.SetTagWithPrefix("efs", tagValue.Key, tagValue.Value)
	}
	properties.Set("Name", e.id)
	properties.Set("ID", e.fsID)
	return properties
}

func (e *EFSMountTarget) String() string {
	return fmt.Sprintf("%s:%s", e.fsID, e.id)
}
