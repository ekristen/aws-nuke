package resources

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/mgn"
	"github.com/aws/aws-sdk-go-v2/service/mgn/types"
	"github.com/aws/smithy-go"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	libtypes "github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const (
	MGNReplicationConfigurationTemplateResource                      = "MGNReplicationConfigurationTemplate"
	mgnReplicationConfigurationTemplateUninitializedAccountException = "UninitializedAccountException"
)

func init() {
	registry.Register(&registry.Registration{
		Name:     MGNReplicationConfigurationTemplateResource,
		Scope:    nuke.Account,
		Resource: &MGNReplicationConfigurationTemplate{},
		Lister:   &MGNReplicationConfigurationTemplateLister{},
	})
}

type MGNReplicationConfigurationTemplateLister struct{}

func (l *MGNReplicationConfigurationTemplateLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := mgn.NewFromConfig(*opts.Config)
	resources := make([]resource.Resource, 0)

	params := &mgn.DescribeReplicationConfigurationTemplatesInput{
		MaxResults: aws.Int32(50),
	}

	for {
		output, err := svc.DescribeReplicationConfigurationTemplates(ctx, params)
		if err != nil {
			var apiErr smithy.APIError
			if errors.As(err, &apiErr) && apiErr.ErrorCode() == mgnReplicationConfigurationTemplateUninitializedAccountException {
				return nil, nil
			}
			return nil, err
		}

		for i := range output.Items {
			template := &output.Items[i]
			mgnTemplate := &MGNReplicationConfigurationTemplate{
				svc:                                svc,
				template:                           template,
				ReplicationConfigurationTemplateID: template.ReplicationConfigurationTemplateID,
				ARN:                                template.Arn,
				StagingAreaSubnetID:                template.StagingAreaSubnetId,
				AssociateDefaultSecurityGroup:      template.AssociateDefaultSecurityGroup,
				BandwidthThrottling:                template.BandwidthThrottling,
				CreatePublicIP:                     template.CreatePublicIP,
				DataPlaneRouting:                   string(template.DataPlaneRouting),
				DefaultLargeStagingDiskType:        string(template.DefaultLargeStagingDiskType),
				EBSEncryption:                      string(template.EbsEncryption),
				EBSEncryptionKeyARN:                template.EbsEncryptionKeyArn,
				ReplicationServerInstanceType:      template.ReplicationServerInstanceType,
				UseDedicatedReplicationServer:      template.UseDedicatedReplicationServer,
				Tags:                               template.Tags,
			}
			resources = append(resources, mgnTemplate)
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type MGNReplicationConfigurationTemplate struct {
	svc      *mgn.Client                             `description:"-"`
	template *types.ReplicationConfigurationTemplate `description:"-"`

	// Exposed properties
	ReplicationConfigurationTemplateID *string           `description:"The unique identifier of the replication configuration template"`
	ARN                                *string           `description:"The ARN of the replication configuration template"`
	StagingAreaSubnetID                *string           `description:"The subnet ID for the staging area"`
	AssociateDefaultSecurityGroup      *bool             `description:"Whether to associate the default security group"`
	BandwidthThrottling                int64             `description:"The bandwidth throttling setting"`
	CreatePublicIP                     *bool             `description:"Whether to create a public IP"`
	DataPlaneRouting                   string            `description:"The data plane routing setting"`
	DefaultLargeStagingDiskType        string            `description:"The default large staging disk type"`
	EBSEncryption                      string            `description:"The EBS encryption setting"`
	EBSEncryptionKeyARN                *string           `description:"The ARN of the EBS encryption key"`
	ReplicationServerInstanceType      *string           `description:"The instance type for the replication server"`
	UseDedicatedReplicationServer      *bool             `description:"Whether to use a dedicated replication server"`
	Tags                               map[string]string `description:"The tags associated with the template"`
}

func (r *MGNReplicationConfigurationTemplate) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteReplicationConfigurationTemplate(ctx, &mgn.DeleteReplicationConfigurationTemplateInput{
		ReplicationConfigurationTemplateID: r.template.ReplicationConfigurationTemplateID,
	})

	return err
}

func (r *MGNReplicationConfigurationTemplate) Properties() libtypes.Properties {
	props := libtypes.NewPropertiesFromStruct(r)
	// TODO(v4): remove
	props.Set("StagingAreaSubnetId", r.StagingAreaSubnetID)
	props.Set("Arn", r.ARN)
	props.Set("EbsEncryption", r.EBSEncryption)
	props.Set("EbsEncryptionKeyArn", r.EBSEncryptionKeyARN)
	return props
}

func (r *MGNReplicationConfigurationTemplate) String() string {
	return *r.ReplicationConfigurationTemplateID
}
