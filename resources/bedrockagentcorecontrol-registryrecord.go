package resources

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	libtypes "github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const BedrockAgentCoreRegistryRecordResource = "BedrockAgentCoreRegistryRecord"

func init() {
	registry.Register(&registry.Registration{
		Name:     BedrockAgentCoreRegistryRecordResource,
		Scope:    nuke.Account,
		Resource: &BedrockAgentCoreRegistryRecord{},
		Lister:   &BedrockAgentCoreRegistryRecordLister{},
	})
}

type BedrockAgentCoreRegistryRecordLister struct {
	BedrockAgentCoreControlLister
}

func (l *BedrockAgentCoreRegistryRecordLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := bedrockagentcorecontrol.NewFromConfig(*opts.Config)
	var resources []resource.Resource

	if !l.IsSupportedRegion(opts.Region.Name) {
		return resources, nil
	}

	registriesParams := &bedrockagentcorecontrol.ListRegistriesInput{
		MaxResults: aws.Int32(100),
	}

	registriesPaginator := bedrockagentcorecontrol.NewListRegistriesPaginator(svc, registriesParams)

	for registriesPaginator.HasMorePages() {
		registriesResp, err := registriesPaginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, reg := range registriesResp.Registries {
			records, err := l.listRecords(ctx, svc, opts, reg.RegistryId, reg.Name)
			if err != nil {
				return nil, err
			}

			resources = append(resources, records...)
		}
	}

	return resources, nil
}

func (l *BedrockAgentCoreRegistryRecordLister) listRecords(ctx context.Context,
	svc *bedrockagentcorecontrol.Client, opts *nuke.ListerOpts,
	registryID, registryName *string) ([]resource.Resource, error) {
	var resources []resource.Resource

	params := &bedrockagentcorecontrol.ListRegistryRecordsInput{
		RegistryId: registryID,
		MaxResults: aws.Int32(100),
	}

	paginator := bedrockagentcorecontrol.NewListRegistryRecordsPaginator(svc, params)

	for paginator.HasMorePages() {
		resp, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, record := range resp.RegistryRecords {
			// Get tags for the registry record
			var tags map[string]string
			tagsResp, err := svc.ListTagsForResource(ctx, &bedrockagentcorecontrol.ListTagsForResourceInput{
				ResourceArn: record.RecordArn,
			})
			if err != nil {
				opts.Logger.Warnf("unable to fetch tags for registry record: %s", *record.RecordArn)
			} else {
				tags = tagsResp.Tags
			}

			resources = append(resources, &BedrockAgentCoreRegistryRecord{
				svc:            svc,
				RecordID:       record.RecordId,
				Name:           record.Name,
				RegistryID:     registryID,
				RegistryName:   registryName,
				DescriptorType: string(record.DescriptorType),
				RecordVersion:  record.RecordVersion,
				Status:         string(record.Status),
				CreatedAt:      record.CreatedAt,
				UpdatedAt:      record.UpdatedAt,
				Tags:           tags,
			})
		}
	}

	return resources, nil
}

type BedrockAgentCoreRegistryRecord struct {
	svc            *bedrockagentcorecontrol.Client
	RecordID       *string
	Name           *string
	RegistryID     *string
	RegistryName   *string
	DescriptorType string
	RecordVersion  *string
	Status         string
	CreatedAt      *time.Time
	UpdatedAt      *time.Time
	Tags           map[string]string
}

func (r *BedrockAgentCoreRegistryRecord) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteRegistryRecord(ctx, &bedrockagentcorecontrol.DeleteRegistryRecordInput{
		RegistryId: r.RegistryID,
		RecordId:   r.RecordID,
	})

	return err
}

func (r *BedrockAgentCoreRegistryRecord) Properties() libtypes.Properties {
	return libtypes.NewPropertiesFromStruct(r)
}

func (r *BedrockAgentCoreRegistryRecord) String() string {
	return *r.RecordID
}
