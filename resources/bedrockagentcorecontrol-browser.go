package resources

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const BedrockAgentCoreBrowserResource = "BedrockAgentCoreBrowser"

func init() {
	registry.Register(&registry.Registration{
		Name:     BedrockAgentCoreBrowserResource,
		Scope:    nuke.Account,
		Resource: &BedrockAgentCoreBrowser{},
		Lister:   &BedrockAgentCoreBrowserLister{},
	})
}

type BedrockAgentCoreBrowserLister struct {
	BedrockAgentCoreControlLister
}

func (l *BedrockAgentCoreBrowserLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := bedrockagentcorecontrol.NewFromConfig(*opts.Config)
	var resources []resource.Resource

	l.SetSupportedRegions(BuiltInToolsSupportedRegions)

	if !l.IsSupportedRegion(opts.Region.Name) {
		return resources, nil
	}

	params := &bedrockagentcorecontrol.ListBrowsersInput{
		MaxResults: aws.Int32(100),
	}

	paginator := bedrockagentcorecontrol.NewListBrowsersPaginator(svc, params)

	for paginator.HasMorePages() {
		resp, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, browser := range resp.BrowserSummaries {
			resources = append(resources, &BedrockAgentCoreBrowser{
				svc:           svc,
				BrowserID:     browser.BrowserId,
				BrowserArn:    browser.BrowserArn,
				Status:        string(browser.Status),
				Description:   browser.Description,
				CreatedAt:     browser.CreatedAt,
				LastUpdatedAt: browser.LastUpdatedAt,
			})
		}
	}

	return resources, nil
}

type BedrockAgentCoreBrowser struct {
	svc           *bedrockagentcorecontrol.Client
	BrowserID     *string
	BrowserArn    *string
	Status        string
	Description   *string
	CreatedAt     *time.Time
	LastUpdatedAt *time.Time
}

func (r *BedrockAgentCoreBrowser) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteBrowser(ctx, &bedrockagentcorecontrol.DeleteBrowserInput{
		BrowserId: r.BrowserID,
	})

	return err
}

func (r *BedrockAgentCoreBrowser) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *BedrockAgentCoreBrowser) String() string {
	return *r.BrowserID
}
