package resources

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/datazone"
	"github.com/aws/aws-sdk-go-v2/service/datazone/types"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	libtypes "github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const DataZoneProjectResource = "DataZoneProject"

func init() {
	registry.Register(&registry.Registration{
		Name:     DataZoneProjectResource,
		Scope:    nuke.Account,
		Resource: &DataZoneProject{},
		Lister:   &DataZoneProjectLister{},
		DependsOn: []string{
			"DataZoneSubscription",
		},
	})
}

type DataZoneProjectLister struct{}

func (l *DataZoneProjectLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := datazone.NewFromConfig(*opts.Config)
	resources := make([]resource.Resource, 0)

	// First, list all domains
	domainParams := &datazone.ListDomainsInput{
		MaxResults: aws.Int32(100),
	}

	domainPaginator := datazone.NewListDomainsPaginator(svc, domainParams)
	for domainPaginator.HasMorePages() {
		domainResp, err := domainPaginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		// For each domain, list its projects
		for _, domain := range domainResp.Items {
			projectParams := &datazone.ListProjectsInput{
				DomainIdentifier: domain.Id,
				MaxResults:       aws.Int32(100),
			}

			projectPaginator := datazone.NewListProjectsPaginator(svc, projectParams)
			for projectPaginator.HasMorePages() {
				projectResp, err := projectPaginator.NextPage(ctx)
				if err != nil {
					return nil, err // Don't swallow errors - fail loudly for SCP denials
				}

				for _, project := range projectResp.Items {
					resources = append(resources, &DataZoneProject{
						svc:           svc,
						DomainID:      domain.Id,
						ID:            project.Id,
						Name:          project.Name,
						CreatedAt:     project.CreatedAt,
						CreatedBy:     project.CreatedBy,
						Description:   project.Description,
						DomainName:    domain.Name,
						ProjectStatus: aws.String(string(project.ProjectStatus)),
					})
				}
			}
		}
	}

	return resources, nil
}

type DataZoneProject struct {
	svc           *datazone.Client
	DomainID      *string
	ID            *string
	Name          *string
	CreatedAt     *time.Time
	CreatedBy     *string
	Description   *string
	DomainName    *string
	ProjectStatus *string
}

func (r *DataZoneProject) Filter() error {
	if r.ProjectStatus != nil {
		switch types.ProjectStatus(*r.ProjectStatus) {
		case types.ProjectStatusDeleteFailed:
			return fmt.Errorf("project is in delete failed state")
		}
	}
	return nil
}

func (r *DataZoneProject) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteProject(ctx, &datazone.DeleteProjectInput{
		DomainIdentifier: r.DomainID,
		Identifier:       r.ID,
	})
	return err
}

func (r *DataZoneProject) HandleWait(ctx context.Context) error {
	resp, err := r.svc.GetProject(ctx, &datazone.GetProjectInput{
		DomainIdentifier: r.DomainID,
		Identifier:       r.ID,
	})
	if err != nil {
		return err
	}

	r.ProjectStatus = aws.String(string(resp.ProjectStatus))

	switch resp.ProjectStatus {
	case types.ProjectStatusDeleteFailed:
		return fmt.Errorf("project deletion failed")
	case types.ProjectStatusDeleting:
		return liberror.ErrWaitResource("project deletion in progress")
	default:
		return liberror.ErrWaitResource(fmt.Sprintf("project status: %s", resp.ProjectStatus))
	}
}

func (r *DataZoneProject) Properties() libtypes.Properties {
	return libtypes.NewPropertiesFromStruct(r)
}
