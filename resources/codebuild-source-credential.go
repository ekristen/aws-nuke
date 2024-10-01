package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/codebuild"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const CodeBuildSourceCredentialResource = "CodeBuildSourceCredential"

func init() {
	registry.Register(&registry.Registration{
		Name:   CodeBuildSourceCredentialResource,
		Scope:  nuke.Account,
		Lister: &CodeBuildSourceCredentialLister{},
	})
}

type CodeBuildSourceCredentialLister struct{}

func (l *CodeBuildSourceCredentialLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := codebuild.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &codebuild.ListSourceCredentialsInput{}

	//This endpoint[1] is not paginated, `SourceCredentialsInfo` doesn't have a `NextToken` field.
	//[1] https://docs.aws.amazon.com/sdk-for-go/api/service/codebuild/#SourceCredentialsInfo
	resp, err := svc.ListSourceCredentials(params)

	if err != nil {
		return nil, err
	}

	for _, credential := range resp.SourceCredentialsInfos {
		resources = append(resources, &CodeBuildSourceCredential{
			svc: svc,
			ARN: credential.Arn,
		})
	}

	return resources, nil
}

type CodeBuildSourceCredential struct {
	svc        *codebuild.CodeBuild
	ARN        *string
	AuthType   *string
	ServerType *string
}

func (r *CodeBuildSourceCredential) Remove(_ context.Context) error {
	_, err := r.svc.DeleteSourceCredentials(&codebuild.DeleteSourceCredentialsInput{Arn: r.ARN})
	return err
}

func (r *CodeBuildSourceCredential) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *CodeBuildSourceCredential) String() string {
	return *r.ARN
}
