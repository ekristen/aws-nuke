package resources

import (
	"context"

	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/service/backup"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

type BackupPlan struct {
	svc  *backup.Backup
	id   string
	name string
	arn  string
	tags map[string]*string
}

const AWSBackupPlanResource = "AWSBackupPlan"

func init() {
	registry.Register(&registry.Registration{
		Name:   AWSBackupPlanResource,
		Scope:  nuke.Account,
		Lister: &AWSBackupPlanLister{},
	})
}

type AWSBackupPlanLister struct{}

func (l *AWSBackupPlanLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := backup.New(opts.Session)
	falseValue := false
	maxBackupsLen := int64(100)
	params := &backup.ListBackupPlansInput{
		IncludeDeleted: &falseValue,
		MaxResults:     &maxBackupsLen, // aws default limit on number of backup plans per account
	}
	resources := make([]resource.Resource, 0)

	for {
		output, err := svc.ListBackupPlans(params)
		if err != nil {
			return nil, err
		}

		for _, plan := range output.BackupPlansList {
			tagsOutput, _ := svc.ListTags(&backup.ListTagsInput{ResourceArn: plan.BackupPlanArn})
			resources = append(resources, &BackupPlan{
				svc:  svc,
				id:   *plan.BackupPlanId,
				name: *plan.BackupPlanName,
				arn:  *plan.BackupPlanArn,
				tags: tagsOutput.Tags,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

func (b *BackupPlan) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("ID", b.id)
	properties.Set("Name", b.name)
	for tagKey, tagValue := range b.tags {
		properties.Set(fmt.Sprintf("tag:%v", tagKey), *tagValue)
	}
	return properties
}

func (b *BackupPlan) Remove(_ context.Context) error {
	_, err := b.svc.DeleteBackupPlan(&backup.DeleteBackupPlanInput{
		BackupPlanId: &b.id,
	})
	return err
}

func (b *BackupPlan) String() string {
	return b.arn
}

func (b *BackupPlan) Filter() error {
	if strings.HasPrefix(b.name, "aws/efs/") {
		return fmt.Errorf("cannot delete EFS automatic backups backup plan")
	}
	return nil
}
