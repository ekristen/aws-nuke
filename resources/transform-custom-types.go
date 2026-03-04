package resources

import "time"

// --- Models ---

// TransformCustomCampaignModel represents a Campaign in the TransformCustom API.
type TransformCustomCampaignModel struct {
	Name                      string    `cbor:"name"`
	Description               string    `cbor:"description"`
	Status                    string    `cbor:"status"`
	TransformationPackageName string    `cbor:"transformationPackageName"`
	CreatedAt                 time.Time `cbor:"createdAt"`
	LastUpdated               time.Time `cbor:"lastUpdated"`
}

// TransformCustomTransformationPackageModel represents a TransformationPackage metadata item.
type TransformCustomTransformationPackageModel struct {
	Version     string    `cbor:"version"`
	Name        string    `cbor:"name"`
	Description string    `cbor:"description"`
	CreatedAt   time.Time `cbor:"createdAt"`
	Verified    bool      `cbor:"verified"`
	Owner       string    `cbor:"owner"`
}

// TransformCustomKnowledgeItemModel represents a KnowledgeItem.
type TransformCustomKnowledgeItemModel struct {
	ID                        string `cbor:"id"`
	TransformationPackageName string `cbor:"transformationPackageName"`
	Title                     string `cbor:"title"`
	Status                    string `cbor:"status"`
}

// TransformCustomCampaignRepositoryModel represents a CampaignRepository.
type TransformCustomCampaignRepositoryModel struct {
	Name        string    `cbor:"name"`
	Status      string    `cbor:"status"`
	LastUpdated time.Time `cbor:"lastUpdated"`
}

// --- Request/Response types ---

// ListCampaigns
type TransformCustomListCampaignsInput struct {
	MaxResults int32  `cbor:"maxResults,omitempty"`
	NextToken  string `cbor:"nextToken,omitempty"`
}

type TransformCustomListCampaignsOutput struct {
	Campaigns []TransformCustomCampaignModel `cbor:"campaigns"`
	NextToken string                         `cbor:"nextToken,omitempty"`
}

// DeleteCampaign
type TransformCustomDeleteCampaignInput struct {
	Name string `cbor:"name"`
}

type TransformCustomDeleteCampaignOutput struct {
	Name string `cbor:"name"`
}

// ListTransformationPackageMetadata
type TransformCustomListTransformationPackageMetadataInput struct {
	MaxResults int32  `cbor:"maxResults,omitempty"`
	NextToken  string `cbor:"nextToken,omitempty"`
	AWSManaged bool   `cbor:"awsManaged,omitempty"`
}

type TransformCustomListTransformationPackageMetadataOutput struct {
	Items     []TransformCustomTransformationPackageModel `cbor:"items"`
	NextToken string                                      `cbor:"nextToken,omitempty"`
}

// DeleteTransformationPackage
type TransformCustomDeleteTransformationPackageInput struct {
	Name string `cbor:"name"`
}

type TransformCustomDeleteTransformationPackageOutput struct {
	Name string `cbor:"name"`
}

// ListKnowledgeItems
type TransformCustomListKnowledgeItemsInput struct {
	TransformationPackageName string `cbor:"transformationPackageName"`
	MaxResults                int32  `cbor:"maxResults,omitempty"`
	NextToken                 string `cbor:"nextToken,omitempty"`
}

type TransformCustomListKnowledgeItemsOutput struct {
	KnowledgeItems []TransformCustomKnowledgeItemModel `cbor:"knowledgeItems"`
	NextToken      string                              `cbor:"nextToken,omitempty"`
}

// DeleteKnowledgeItem
type TransformCustomDeleteKnowledgeItemInput struct {
	ID                        string `cbor:"id"`
	TransformationPackageName string `cbor:"transformationPackageName"`
}

type TransformCustomDeleteKnowledgeItemOutput struct {
	ID string `cbor:"id"`
}

// ListCampaignRepositories
type TransformCustomListCampaignRepositoriesInput struct {
	Name       string `cbor:"name"`
	MaxResults int32  `cbor:"maxResults,omitempty"`
	NextToken  string `cbor:"nextToken,omitempty"`
}

type TransformCustomListCampaignRepositoriesOutput struct {
	Repositories []TransformCustomCampaignRepositoryModel `cbor:"repositories"`
	NextToken    string                                   `cbor:"nextToken,omitempty"`
}
