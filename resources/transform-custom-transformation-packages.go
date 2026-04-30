package resources

import (
	"context"
	"fmt"
	"time"

	"github.com/gotidy/ptr"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const TransformCustomTransformationPackageResource = "TransformCustomTransformationPackage"

func init() {
	registry.Register(&registry.Registration{
		Name:     TransformCustomTransformationPackageResource,
		Scope:    nuke.Account,
		Resource: &TransformCustomTransformationPackage{},
		Lister:   &TransformCustomTransformationPackageLister{},
	})
}

type TransformCustomTransformationPackageLister struct {
	svc TransformCustomAPI
}

func (l *TransformCustomTransformationPackageLister) List(
	ctx context.Context, o interface{},
) ([]resource.Resource, error) {
	if l.svc == nil {
		opts := o.(*nuke.ListerOpts)
		l.svc = NewTransformCustomClient(opts.Config)
	}

	var resources []resource.Resource

	params := &TransformCustomListTransformationPackageMetadataInput{
		MaxResults: 100,
	}

	for {
		resp, err := l.svc.ListTransformationPackageMetadata(ctx, params)
		if err != nil {
			return nil, err
		}

		for _, item := range resp.Items {
			resources = append(resources, &TransformCustomTransformationPackage{
				svc:         l.svc,
				Name:        ptr.String(item.Name),
				Version:     ptr.String(item.Version),
				Description: ptr.String(item.Description),
				CreatedAt:   ptr.Time(item.CreatedAt),
				Verified:    ptr.Bool(item.Verified),
				Owner:       ptr.String(item.Owner),
			})
		}

		if resp.NextToken == "" {
			break
		}

		params.NextToken = resp.NextToken
	}

	return resources, nil
}

type TransformCustomTransformationPackage struct {
	svc         TransformCustomAPI
	Name        *string
	Version     *string
	Description *string
	CreatedAt   *time.Time
	Verified    *bool
	Owner       *string
}

func (r *TransformCustomTransformationPackage) Filter() error {
	if ptr.ToString(r.Owner) == "AWS" {
		return fmt.Errorf("cannot delete AWS-managed transformation package")
	}
	return nil
}

func (r *TransformCustomTransformationPackage) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteTransformationPackage(ctx, &TransformCustomDeleteTransformationPackageInput{
		Name: ptr.ToString(r.Name),
	})
	return err
}

func (r *TransformCustomTransformationPackage) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *TransformCustomTransformationPackage) String() string {
	return ptr.ToString(r.Name)
}
