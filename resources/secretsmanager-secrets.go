package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/secretsmanager"

	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const SecretsManagerSecretResource = "SecretsManagerSecret"

func init() {
	resource.Register(resource.Registration{
		Name:   SecretsManagerSecretResource,
		Scope:  nuke.Account,
		Lister: &SecretsManagerSecretLister{},
	})
}

type SecretsManagerSecretLister struct{}

func (l *SecretsManagerSecretLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := secretsmanager.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &secretsmanager.ListSecretsInput{
		MaxResults: aws.Int64(100),
	}

	for {
		output, err := svc.ListSecrets(params)
		if err != nil {
			return nil, err
		}

		for _, secrets := range output.SecretList {
			resources = append(resources, &SecretsManagerSecret{
				svc:  svc,
				ARN:  secrets.ARN,
				tags: secrets.Tags,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type SecretsManagerSecret struct {
	svc  *secretsmanager.SecretsManager
	ARN  *string
	tags []*secretsmanager.Tag
}

func (f *SecretsManagerSecret) Remove(_ context.Context) error {
	_, err := f.svc.DeleteSecret(&secretsmanager.DeleteSecretInput{
		SecretId:                   f.ARN,
		ForceDeleteWithoutRecovery: aws.Bool(true),
	})

	return err
}

func (f *SecretsManagerSecret) Properties() types.Properties {
	properties := types.NewProperties()
	for _, tagValue := range f.tags {
		properties.SetTag(tagValue.Key, tagValue.Value)
	}
	return properties
}

func (f *SecretsManagerSecret) String() string {
	return *f.ARN
}
