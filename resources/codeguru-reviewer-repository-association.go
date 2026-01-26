package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/codegurureviewer" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const CodeGuruReviewerRepositoryAssociationResource = "CodeGuruReviewerRepositoryAssociation"

func init() {
	registry.Register(&registry.Registration{
		Name:     CodeGuruReviewerRepositoryAssociationResource,
		Scope:    nuke.Account,
		Resource: &CodeGuruReviewerRepositoryAssociation{},
		Lister:   &CodeGuruReviewerRepositoryAssociationLister{},
	})
}

type CodeGuruReviewerRepositoryAssociationLister struct{}

func (l *CodeGuruReviewerRepositoryAssociationLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	var resources []resource.Resource

	svc := codegurureviewer.New(opts.Session)

	params := &codegurureviewer.ListRepositoryAssociationsInput{}

	for {
		resp, err := svc.ListRepositoryAssociations(params)
		if err != nil {
			return nil, err
		}

		for _, association := range resp.RepositoryAssociationSummaries {
			resources = append(resources, &CodeGuruReviewerRepositoryAssociation{
				svc:            svc,
				AssociationARN: association.AssociationArn,
				AssociationID:  association.AssociationId,
				Name:           association.Name,
				Owner:          association.Owner,
				ProviderType:   association.ProviderType,
			})
		}

		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return resources, nil
}

type CodeGuruReviewerRepositoryAssociation struct {
	svc            *codegurureviewer.CodeGuruReviewer
	AssociationARN *string
	AssociationID  *string
	Name           *string
	Owner          *string
	ProviderType   *string
}

func (r *CodeGuruReviewerRepositoryAssociation) Remove(_ context.Context) error {
	_, err := r.svc.DisassociateRepository(&codegurureviewer.DisassociateRepositoryInput{
		AssociationArn: r.AssociationARN,
	})
	return err
}

func (r *CodeGuruReviewerRepositoryAssociation) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *CodeGuruReviewerRepositoryAssociation) String() string {
	return *r.Name
}
