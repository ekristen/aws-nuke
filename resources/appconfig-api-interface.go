package resources

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/appconfig"
)

// AppConfigAPI defines the interface for the AppConfig API operations used by the AppConfig extension resources.
// Defined for dependency injection and test mocking.
type AppConfigAPI interface {
	ListExtensions(ctx context.Context, params *appconfig.ListExtensionsInput,
		optFns ...func(*appconfig.Options)) (*appconfig.ListExtensionsOutput, error)
	DeleteExtension(ctx context.Context, params *appconfig.DeleteExtensionInput,
		optFns ...func(*appconfig.Options)) (*appconfig.DeleteExtensionOutput, error)
	ListExtensionAssociations(ctx context.Context, params *appconfig.ListExtensionAssociationsInput,
		optFns ...func(*appconfig.Options)) (*appconfig.ListExtensionAssociationsOutput, error)
	DeleteExtensionAssociation(ctx context.Context, params *appconfig.DeleteExtensionAssociationInput,
		optFns ...func(*appconfig.Options)) (*appconfig.DeleteExtensionAssociationOutput, error)
}
