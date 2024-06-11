package resources

import (
	"context"
	"strings"

	"github.com/gotidy/ptr"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/secretsmanager"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const SecretsManagerSecretResource = "SecretsManagerSecret"

func init() {
	registry.Register(&registry.Registration{
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

		for _, secret := range output.SecretList {
			replica := false
			var primarySvc *secretsmanager.SecretsManager
			if opts.Region.Name != *secret.PrimaryRegion {
				replica = true

				primaryCfg := opts.Session.Copy(&aws.Config{
					Region: aws.String(*secret.PrimaryRegion),
				})

				primarySvc = secretsmanager.New(primaryCfg)
			}

			resources = append(resources, &SecretsManagerSecret{
				svc:           svc,
				primarySvc:    primarySvc,
				region:        ptr.String(opts.Region.Name),
				ARN:           secret.ARN,
				Name:          secret.Name,
				PrimaryRegion: secret.PrimaryRegion,
				Replica:       replica,
				tags:          secret.Tags,
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
	svc            *secretsmanager.SecretsManager
	primarySvc     *secretsmanager.SecretsManager
	region         *string
	ARN            *string
	Name           *string
	PrimaryRegion  *string
	Replica        bool
	ReplicaRegions []*string
	tags           []*secretsmanager.Tag
}

// ParentARN returns the ARN of the parent secret by doing a string replace of the region. For example, if the region
// is us-west-2 and the primary region is us-east-1, the ARN will be replaced with us-east-1. This will allow for the
// RemoveRegionsFromReplication call to work properly.
func (r *SecretsManagerSecret) ParentARN() *string {
	return ptr.String(strings.ReplaceAll(*r.ARN, *r.region, *r.PrimaryRegion))
}

func (r *SecretsManagerSecret) Remove(_ context.Context) error {
	if r.Replica {
		_, err := r.primarySvc.RemoveRegionsFromReplication(&secretsmanager.RemoveRegionsFromReplicationInput{
			SecretId:             r.ParentARN(),
			RemoveReplicaRegions: []*string{r.region},
		})

		return err
	}

	_, err := r.svc.DeleteSecret(&secretsmanager.DeleteSecretInput{
		SecretId:                   r.ARN,
		ForceDeleteWithoutRecovery: aws.Bool(true),
	})

	return err
}

func (r *SecretsManagerSecret) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("PrimaryRegion", r.PrimaryRegion)
	properties.Set("Replica", r.Replica)
	properties.Set("Name", r.Name)
	properties.Set("ARN", r.ARN)
	for _, tagValue := range r.tags {
		properties.SetTag(tagValue.Key, tagValue.Value)
	}
	return properties
}

// TODO(v4): change the return value to the name of the resource instead of the ARN
func (r *SecretsManagerSecret) String() string {
	return *r.ARN
}
