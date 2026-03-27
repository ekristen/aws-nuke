package resources

import (
	"context"

	"github.com/gotidy/ptr"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const TransformCustomKnowledgeItemResource = "TransformCustomKnowledgeItem"

func init() {
	registry.Register(&registry.Registration{
		Name:     TransformCustomKnowledgeItemResource,
		Scope:    nuke.Account,
		Resource: &TransformCustomKnowledgeItem{},
		Lister:   &TransformCustomKnowledgeItemLister{},
	})
}

type TransformCustomKnowledgeItemLister struct {
	svc TransformCustomAPI
}

func (l *TransformCustomKnowledgeItemLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	if l.svc == nil {
		opts := o.(*nuke.ListerOpts)
		l.svc = NewTransformCustomClient(opts.Config)
	}

	// First, list all transformation packages to enumerate knowledge items across them
	pkgParams := &TransformCustomListTransformationPackageMetadataInput{
		MaxResults: 100,
	}

	var pkgNames []string

	for {
		pkgResp, err := l.svc.ListTransformationPackageMetadata(ctx, pkgParams)
		if err != nil {
			return nil, err
		}

		for _, item := range pkgResp.Items {
			pkgNames = append(pkgNames, item.Name)
		}

		if pkgResp.NextToken == "" {
			break
		}

		pkgParams.NextToken = pkgResp.NextToken
	}

	var resources []resource.Resource

	for _, pkgName := range pkgNames {
		params := &TransformCustomListKnowledgeItemsInput{
			TransformationPackageName: pkgName,
			MaxResults:                100,
		}

		for {
			resp, err := l.svc.ListKnowledgeItems(ctx, params)
			if err != nil {
				return nil, err
			}

			for _, ki := range resp.KnowledgeItems {
				resources = append(resources, &TransformCustomKnowledgeItem{
					svc:                       l.svc,
					ID:                        ptr.String(ki.ID),
					TransformationPackageName: ptr.String(ki.TransformationPackageName),
					Title:                     ptr.String(ki.Title),
					Status:                    ptr.String(ki.Status),
				})
			}

			if resp.NextToken == "" {
				break
			}

			params.NextToken = resp.NextToken
		}
	}

	return resources, nil
}

type TransformCustomKnowledgeItem struct {
	svc                       TransformCustomAPI
	ID                        *string
	TransformationPackageName *string
	Title                     *string
	Status                    *string
}

func (r *TransformCustomKnowledgeItem) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteKnowledgeItem(ctx, &TransformCustomDeleteKnowledgeItemInput{
		ID:                        ptr.ToString(r.ID),
		TransformationPackageName: ptr.ToString(r.TransformationPackageName),
	})
	return err
}

func (r *TransformCustomKnowledgeItem) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *TransformCustomKnowledgeItem) String() string {
	return ptr.ToString(r.ID)
}
