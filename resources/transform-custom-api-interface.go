package resources

import "context"

// TransformCustomAPI is the interface for the TransformCustom service API.
// Used for dependency injection and test mocking.
type TransformCustomAPI interface {
	ListCampaigns(ctx context.Context, params *TransformCustomListCampaignsInput) (
		*TransformCustomListCampaignsOutput, error)
	DeleteCampaign(ctx context.Context, params *TransformCustomDeleteCampaignInput) (
		*TransformCustomDeleteCampaignOutput, error)
	ListTransformationPackageMetadata(ctx context.Context, params *TransformCustomListTransformationPackageMetadataInput) (
		*TransformCustomListTransformationPackageMetadataOutput, error)
	DeleteTransformationPackage(ctx context.Context, params *TransformCustomDeleteTransformationPackageInput) (
		*TransformCustomDeleteTransformationPackageOutput, error)
	ListKnowledgeItems(ctx context.Context, params *TransformCustomListKnowledgeItemsInput) (
		*TransformCustomListKnowledgeItemsOutput, error)
	DeleteKnowledgeItem(ctx context.Context, params *TransformCustomDeleteKnowledgeItemInput) (
		*TransformCustomDeleteKnowledgeItemOutput, error)
	ListCampaignRepositories(ctx context.Context, params *TransformCustomListCampaignRepositoriesInput) (
		*TransformCustomListCampaignRepositoriesOutput, error)
}
