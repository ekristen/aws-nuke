package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/elbv2"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

// Register the resource type name that will be used in the config include/exclude lists.
const ELBv2TrustStoreResource = "ELBv2TrustStore"

func init() {
	registry.Register(&registry.Registration{
		Name:     ELBv2TrustStoreResource,
		Scope:    nuke.Account,
		Resource: &ELBv2TrustStore{},
		Lister:   &ELBv2TrustStoreLister{},
	})
}

type ELBv2TrustStoreLister struct{}

// List discovers all ELBv2 Trust Stores in the current region for the
// AWS account and returns them as libnuke resources.
func (l *ELBv2TrustStoreLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := elbv2.New(opts.Session)
	resources := make([]resource.Resource, 0)

	// DescribeTrustStores supports paginated responses.
	err := svc.DescribeTrustStoresPages(&elbv2.DescribeTrustStoresInput{},
		func(page *elbv2.DescribeTrustStoresOutput, _ bool) bool {
			for _, ts := range page.TrustStores {
				resources = append(resources, &ELBv2TrustStore{
					svc:        svc,
					trustStore: ts,
				})
			}
			// Always continue until the API indicates the last page.
			return true
		})
	if err != nil {
		return nil, err
	}

	return resources, nil
}

type ELBv2TrustStore struct {
	svc        *elbv2.ELBV2
	trustStore *elbv2.TrustStore
}

// Remove deletes the trust store.
func (e *ELBv2TrustStore) Remove(_ context.Context) error {
	_, err := e.svc.DeleteTrustStore(&elbv2.DeleteTrustStoreInput{
		TrustStoreArn: e.trustStore.TrustStoreArn,
	})
	return err
}

// Properties returns structured information about the trust store to be used
// in reporting and filtering.
func (e *ELBv2TrustStore) Properties() types.Properties {
	props := types.NewProperties().
		Set("ARN", e.trustStore.TrustStoreArn).
		Set("Name", e.trustStore.Name).
		Set("Status", e.trustStore.Status)

	if e.trustStore.NumberOfCaCertificates != nil {
		props.Set("NumberOfCaCertificates", e.trustStore.NumberOfCaCertificates)
	}
	if e.trustStore.TotalRevokedEntries != nil {
		props.Set("TotalRevokedEntries", e.trustStore.TotalRevokedEntries)
	}

	return props
}

// String returns a human-readable identifier for the trust store.
func (e *ELBv2TrustStore) String() string {
	if e.trustStore.Name != nil {
		return *e.trustStore.Name
	}
	if e.trustStore.TrustStoreArn != nil {
		return *e.trustStore.TrustStoreArn
	}
	return "ELBv2 TrustStore"
}
