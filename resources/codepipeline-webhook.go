package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/codepipeline"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const CodePipelineWebhookResource = "CodePipelineWebhook"

func init() {
	registry.Register(&registry.Registration{
		Name:     CodePipelineWebhookResource,
		Scope:    nuke.Account,
		Resource: &CodePipelineWebhook{},
		Lister:   &CodePipelineWebhookLister{},
	})
}

type CodePipelineWebhookLister struct{}

func (l *CodePipelineWebhookLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	var resources []resource.Resource

	svc := codepipeline.New(opts.Session)

	params := &codepipeline.ListWebhooksInput{}

	for {
		resp, err := svc.ListWebhooks(params)
		if err != nil {
			return nil, err
		}

		for _, webHooks := range resp.Webhooks {
			resources = append(resources, &CodePipelineWebhook{
				svc:  svc,
				Name: webHooks.Definition.Name,
			})
		}

		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return resources, nil
}

type CodePipelineWebhook struct {
	svc  *codepipeline.CodePipeline
	Name *string
}

func (r *CodePipelineWebhook) Remove(_ context.Context) error {
	_, err := r.svc.DeleteWebhook(&codepipeline.DeleteWebhookInput{
		Name: r.Name,
	})

	return err
}

func (r *CodePipelineWebhook) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *CodePipelineWebhook) String() string {
	return *r.Name
}
