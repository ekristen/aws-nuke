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
			// Get tags for the browser
			var tags map[string]string
			if browser.BrowserArn != nil {
				tagsResp, err := svc.ListTagsForResource(ctx, &bedrockagentcorecontrol.ListTagsForResourceInput{
					ResourceArn: browser.BrowserArn,
				})
				if err != nil {
					opts.Logger.Warnf("unable to fetch tags for browser: %s", *browser.BrowserArn)
				} else {
					tags = tagsResp.Tags
				}
			}

			resources = append(resources, &BedrockAgentCoreBrowser{
				svc:           svc,
				ID:            browser.BrowserId,
				Name:          browser.Name,
				Status:        string(browser.Status),
				CreatedAt:     browser.CreatedAt,
				LastUpdatedAt: browser.LastUpdatedAt,
				Tags:          tags,
			})
		}
	}

	return resources, nil
}

type BedrockAgentCoreBrowser struct {
	svc           *bedrockagentcorecontrol.Client
	ID            *string
	Name          *string
	Status        string
	CreatedAt     *time.Time
	LastUpdatedAt *time.Time
	Tags          map[string]string
}

func (r *BedrockAgentCoreBrowser) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteBrowser(ctx, &bedrockagentcorecontrol.DeleteBrowserInput{
		BrowserId: r.ID,
	})

	return err
}

func (r *BedrockAgentCoreBrowser) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *BedrockAgentCoreBrowser) String() string {
	return *r.Name
}
