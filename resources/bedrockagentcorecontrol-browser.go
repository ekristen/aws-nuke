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

const BedrockBrowserResource = "BedrockBrowser"

func init() {
	registry.Register(&registry.Registration{
		Name:     BedrockBrowserResource,
		Scope:    nuke.Account,
		Resource: &BedrockBrowser{},
		Lister:   &BedrockBrowserLister{},
	})
}

type BedrockBrowserLister struct{}

func (l *BedrockBrowserLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := bedrockagentcorecontrol.NewFromConfig(*opts.Config)
	var resources []resource.Resource

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
			resources = append(resources, &BedrockBrowser{
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

type BedrockBrowser struct {
	svc           *bedrockagentcorecontrol.Client
	BrowserID     *string
	BrowserArn    *string
	Status        string
	Description   *string
	CreatedAt     *time.Time
	LastUpdatedAt *time.Time
}

func (r *BedrockBrowser) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteBrowser(ctx, &bedrockagentcorecontrol.DeleteBrowserInput{
		BrowserId: r.BrowserID,
	})

	return err
}

func (r *BedrockBrowser) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *BedrockBrowser) String() string {
	return *r.BrowserID
}
