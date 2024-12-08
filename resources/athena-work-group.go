package resources

import (
	"context"
	"errors"
	"fmt"

	"github.com/gotidy/ptr"
	"github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/athena"
	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const AthenaWorkGroupResource = "AthenaWorkGroup"

func init() {
	registry.Register(&registry.Registration{
		Name:     AthenaWorkGroupResource,
		Scope:    nuke.Account,
		Resource: &AthenaWorkGroup{},
		Lister:   &AthenaWorkGroupLister{},
	})
}

type AthenaWorkGroupLister struct{}

func (l *AthenaWorkGroupLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := athena.New(opts.Session)
	resources := make([]resource.Resource, 0)

	// List WorkGroup
	var workgroupNames []*string
	err := svc.ListWorkGroupsPages(
		&athena.ListWorkGroupsInput{},
		func(page *athena.ListWorkGroupsOutput, lastPage bool) bool {
			for _, workgroup := range page.WorkGroups {
				workgroupNames = append(workgroupNames, workgroup.Name)
			}
			return true
		},
	)
	if err != nil {
		return nil, err
	}

	// Create AthenaWorkGroup resource objects
	for _, name := range workgroupNames {
		resources = append(resources, &AthenaWorkGroup{
			svc:  svc,
			name: name,
			// The GetWorkGroup API doesn't return an ARN,
			// so we need to construct one ourselves
			arn: aws.String(fmt.Sprintf(
				"arn:aws:athena:%s:%s:workgroup/%s",
				opts.Region.Name, *opts.AccountID, *name,
			)),
		})
	}

	return resources, err
}

type AthenaWorkGroup struct {
	svc  *athena.Athena
	name *string
	arn  *string
}

func (r *AthenaWorkGroup) Remove(_ context.Context) error {
	// Primary WorkGroup cannot be deleted,
	// but we can reset it to r clean state
	if *r.name == "primary" {
		// TODO: pass logger via ListerOpts instead of using global
		logrus.Info("Primary Athena WorkGroup may not be deleted. Resetting configuration only.")

		// Reset the configuration to its default state
		_, err := r.svc.UpdateWorkGroup(&athena.UpdateWorkGroupInput{
			// See https://docs.aws.amazon.com/athena/latest/APIReference/API_WorkGroupConfigurationUpdates.html
			// for documented defaults
			ConfigurationUpdates: &athena.WorkGroupConfigurationUpdates{
				EnforceWorkGroupConfiguration:    aws.Bool(false),
				PublishCloudWatchMetricsEnabled:  aws.Bool(false),
				RemoveBytesScannedCutoffPerQuery: aws.Bool(true),
				RequesterPaysEnabled:             aws.Bool(false),
				ResultConfigurationUpdates: &athena.ResultConfigurationUpdates{
					RemoveEncryptionConfiguration: aws.Bool(true),
					RemoveOutputLocation:          aws.Bool(true),
				},
			},
			Description: aws.String(""),
			WorkGroup:   r.name,
		})
		if err != nil {
			return err
		}

		// Remove any tags
		wgTagsRes, err := r.svc.ListTagsForResource(&athena.ListTagsForResourceInput{
			ResourceARN: r.arn,
		})
		if err != nil {
			return err
		}

		var tagKeys []*string
		for _, tag := range wgTagsRes.Tags {
			tagKeys = append(tagKeys, tag.Key)
		}

		_, err = r.svc.UntagResource(&athena.UntagResourceInput{
			ResourceARN: r.arn,
			TagKeys:     tagKeys,
		})
		if err != nil {
			return err
		}

		return nil
	}

	_, err := r.svc.DeleteWorkGroup(&athena.DeleteWorkGroupInput{
		RecursiveDeleteOption: aws.Bool(true),
		WorkGroup:             r.name,
	})

	return err
}

func (r *AthenaWorkGroup) Filter() error {
	// If this is the primary work group,
	// check if it's already had its configuration reset
	if *r.name == "primary" {
		// Get workgroup configuration
		wgConfigRes, err := r.svc.GetWorkGroup(&athena.GetWorkGroupInput{
			WorkGroup: r.name,
		})
		if err != nil {
			return err
		}

		// Get workgroup tags
		wgTagsRes, err := r.svc.ListTagsForResource(&athena.ListTagsForResourceInput{
			ResourceARN: r.arn,
		})
		if err != nil {
			return err
		}

		// If the workgroup is already in r "clean" state, then
		// don't add it to our plan
		wgConfig := wgConfigRes.WorkGroup.Configuration
		isCleanConfig := wgConfig.BytesScannedCutoffPerQuery == nil &&
			!ptr.ToBool(wgConfig.EnforceWorkGroupConfiguration) &&
			!ptr.ToBool(wgConfig.PublishCloudWatchMetricsEnabled) &&
			!ptr.ToBool(wgConfig.RequesterPaysEnabled) &&
			*wgConfig.ResultConfiguration == athena.ResultConfiguration{} &&
			len(wgTagsRes.Tags) == 0

		if isCleanConfig {
			return errors.New("cannot delete primary athena work group")
		}
	}
	return nil
}

func (r *AthenaWorkGroup) Properties() types.Properties {
	return types.NewProperties().
		Set("Name", *r.name).
		Set("ARN", *r.arn)
}

func (r *AthenaWorkGroup) String() string {
	return *r.name
}
