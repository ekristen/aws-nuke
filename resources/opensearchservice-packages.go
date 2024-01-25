package resources

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/service/opensearchservice"

	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const OSPackageResource = "OSPackage"

func init() {
	resource.Register(&resource.Registration{
		Name:   OSPackageResource,
		Scope:  nuke.Account,
		Lister: &OSPackageLister{},
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
				packageID:   pkg.PackageID,
				packageName: pkg.PackageName,
				createdTime: pkg.CreatedAt,
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
	packageID   *string
	packageName *string
	createdTime *time.Time
}

func (o *OSPackage) Filter() error {
	if strings.HasPrefix(*o.packageID, "G") {
		return fmt.Errorf("cannot delete default opensearch packages")
	}
	return nil
}

func (o *OSPackage) Remove(_ context.Context) error {
	_, err := o.svc.DeletePackage(&opensearchservice.DeletePackageInput{
		PackageID: o.packageID,
	})

	return err
}

func (o *OSPackage) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("PackageID", o.packageID)
	properties.Set("PackageName", o.packageName)
	properties.Set("CreatedTime", o.createdTime.Format(time.RFC3339))
	return properties
}

func (o *OSPackage) String() string {
	return *o.packageID
}
