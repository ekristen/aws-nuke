package resources

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/fxamacker/cbor/v2"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
)

const (
	transformCustomServiceName = "transform-custom"
	transformCustomAPIVersion  = "2022-07-26"
)

// TransformCustomClient implements TransformCustomAPI using Smithy RPC v2 CBOR over HTTP.
type TransformCustomClient struct {
	cfg        aws.Config
	endpoint   string
	httpClient *http.Client
}

// NewTransformCustomClient creates a new client for the TransformCustom API.
func NewTransformCustomClient(cfg *aws.Config) *TransformCustomClient {
	return &TransformCustomClient{
		cfg:        *cfg,
		endpoint:   fmt.Sprintf("https://%s.%s.api.aws", transformCustomServiceName, cfg.Region),
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *TransformCustomClient) invoke(ctx context.Context, operation string, input, output interface{}) error {
	body, err := cbor.Marshal(input)
	if err != nil {
		return fmt.Errorf("cbor marshal: %w", err)
	}

	url := fmt.Sprintf("%s/service/TransformCustom/operation/%s", c.endpoint, operation)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}

	req.Header.Set("Content-Type", "application/cbor")
	req.Header.Set("Accept", "application/cbor")
	req.Header.Set("smithy-protocol", "rpc-v2-cbor")

	creds, err := c.cfg.Credentials.Retrieve(ctx)
	if err != nil {
		return fmt.Errorf("retrieve credentials: %w", err)
	}

	hash := sha256.Sum256(body)
	payloadHash := hex.EncodeToString(hash[:])

	signer := v4.NewSigner()
	err = signer.SignHTTP(ctx, creds, req, payloadHash, transformCustomServiceName, c.cfg.Region, time.Now())
	if err != nil {
		return fmt.Errorf("sign request: %w", err)
	}

	resp, err := c.httpClient.Do(req) //nolint:gosec // URL built from trusted aws.Config endpoint
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("TransformCustom %s: HTTP %d: %s", operation, resp.StatusCode, string(respBody))
	}

	if err := cbor.Unmarshal(respBody, output); err != nil {
		return fmt.Errorf("cbor unmarshal: %w", err)
	}

	return nil
}

func (c *TransformCustomClient) ListCampaigns(
	ctx context.Context, params *TransformCustomListCampaignsInput,
) (*TransformCustomListCampaignsOutput, error) {
	out := &TransformCustomListCampaignsOutput{}
	err := c.invoke(ctx, "ListCampaigns", params, out)
	return out, err
}

func (c *TransformCustomClient) DeleteCampaign(
	ctx context.Context, params *TransformCustomDeleteCampaignInput,
) (*TransformCustomDeleteCampaignOutput, error) {
	out := &TransformCustomDeleteCampaignOutput{}
	err := c.invoke(ctx, "DeleteCampaign", params, out)
	return out, err
}

func (c *TransformCustomClient) ListTransformationPackageMetadata(
	ctx context.Context, params *TransformCustomListTransformationPackageMetadataInput,
) (*TransformCustomListTransformationPackageMetadataOutput, error) {
	out := &TransformCustomListTransformationPackageMetadataOutput{}
	err := c.invoke(ctx, "ListTransformationPackageMetadata", params, out)
	return out, err
}

func (c *TransformCustomClient) DeleteTransformationPackage(
	ctx context.Context, params *TransformCustomDeleteTransformationPackageInput,
) (*TransformCustomDeleteTransformationPackageOutput, error) {
	out := &TransformCustomDeleteTransformationPackageOutput{}
	err := c.invoke(ctx, "DeleteTransformationPackage", params, out)
	return out, err
}

func (c *TransformCustomClient) ListKnowledgeItems(
	ctx context.Context, params *TransformCustomListKnowledgeItemsInput,
) (*TransformCustomListKnowledgeItemsOutput, error) {
	out := &TransformCustomListKnowledgeItemsOutput{}
	err := c.invoke(ctx, "ListKnowledgeItems", params, out)
	return out, err
}

func (c *TransformCustomClient) DeleteKnowledgeItem(
	ctx context.Context, params *TransformCustomDeleteKnowledgeItemInput,
) (*TransformCustomDeleteKnowledgeItemOutput, error) {
	out := &TransformCustomDeleteKnowledgeItemOutput{}
	err := c.invoke(ctx, "DeleteKnowledgeItem", params, out)
	return out, err
}

func (c *TransformCustomClient) ListCampaignRepositories(
	ctx context.Context, params *TransformCustomListCampaignRepositoriesInput,
) (*TransformCustomListCampaignRepositoriesOutput, error) {
	out := &TransformCustomListCampaignRepositoriesOutput{}
	err := c.invoke(ctx, "ListCampaignRepositories", params, out)
	return out, err
}
