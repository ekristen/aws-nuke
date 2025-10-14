package resources

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/service/opensearchservice" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const OSPackageResource = "OSPackage"

func init() {
	registry.Register(&registry.Registration{
		Name:     OSPackageResource,
		Scope:    nuke.Account,
		Resource: &OSPackage{},
		Lister:   &OSPackageLister{},
	})
}

type OSPackageLister struct{}

func (l *OSPackageLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := opensearchservice.New(opts.Session)
	resources := make([]resource.Resource, 0)
	var nextToken *string

	for {
		params := &opensearchservice.DescribePackagesInput{
			NextToken: nextToken,
		}
		listResp, err := svc.DescribePackages(params)
		if err != nil {
			return nil, err
		}

		for _, pkg := range listResp.PackageDetailsList {
			resources = append(resources, &OSPackage{
				svc:         svc,
				PackageID:   pkg.PackageID,
				PackageName: pkg.PackageName,
				PackageType: pkg.PackageType,
				CreatedTime: pkg.CreatedAt,
			})
		}

		// Check if there are more results
		if listResp.NextToken == nil {
			break // No more results, exit the loop
		}

		// Set the nextToken for the next iteration
		nextToken = listResp.NextToken
	}

	return resources, nil
}

type OSPackage struct {
	svc         *opensearchservice.OpenSearchService
	PackageID   *string
	PackageName *string
	PackageType *string
	CreatedTime *time.Time
}

func (r *OSPackage) Filter() error {
	if strings.HasPrefix(*r.PackageID, "G") {
		return fmt.Errorf("cannot delete default opensearch packages")
	}

	if *r.PackageType == "ZIP-PLUGIN" {
		return fmt.Errorf("cannot delete opensearch package plugin")
	}
	return nil
}

func (r *OSPackage) Remove(_ context.Context) error {
	_, err := r.svc.DeletePackage(&opensearchservice.DeletePackageInput{
		PackageID: r.PackageID,
	})

	return err
}

func (r *OSPackage) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *OSPackage) String() string {
	return *r.PackageID
}
