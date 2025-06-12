package resources

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/gotidy/ptr"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-sdk-go/service/secretsmanager/secretsmanageriface"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const SecretsManagerSecretResource = "SecretsManagerSecret"

var managedRegex = regexp.MustCompile("^([a-z-]+)!.*$")
var errAWSManaged = errors.New("cannot delete AWS managed secret")

func init() {
	registry.Register(&registry.Registration{
		Name:     SecretsManagerSecretResource,
		Scope:    nuke.Account,
		Resource: &SecretsManagerSecret{},
		Lister:   &SecretsManagerSecretLister{},
	})
}

type SecretsManagerSecretLister struct {
	mockSvc secretsmanageriface.SecretsManagerAPI
}

func (l *SecretsManagerSecretLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	var svc secretsmanageriface.SecretsManagerAPI
	if l.mockSvc != nil {
		svc = l.mockSvc
	} else {
		svc = secretsmanager.New(opts.Session)
	}

	resources := make([]resource.Resource, 0)

	params := &secretsmanager.ListSecretsInput{
		MaxResults:             aws.Int64(100),
		IncludePlannedDeletion: aws.Bool(true),
	}

	for {
		output, err := svc.ListSecrets(params)
		if err != nil {
			return nil, err
		}

		for _, secret := range output.SecretList {
			replica := false
			var primarySvc *secretsmanager.SecretsManager

			// Note: if primary region is not set, then the secret is not a replica
			primaryRegion := ptr.ToString(secret.PrimaryRegion)
			if primaryRegion != "" && opts.Region.Name != primaryRegion {
				replica = true

				primaryCfg := opts.Session.Copy(&aws.Config{
					Region: secret.PrimaryRegion,
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
	svc            secretsmanageriface.SecretsManagerAPI
	primarySvc     secretsmanageriface.SecretsManagerAPI
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

func (r *SecretsManagerSecret) Filter() error {
	if managedRegex.MatchString(*r.Name) {
		return errAWSManaged
	}

	for _, tag := range r.tags {
		if *tag.Key == "aws:secretsmanager:owningService" {
			return errAWSManaged
		}
	}

	return nil
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
