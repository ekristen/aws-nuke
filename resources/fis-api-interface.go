package resources

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/fis"
)

// FISAPI defines the interface for AWS Fault Injection Service API operations.
// Defined for dependency injection and test mocking.
type FISAPI interface {
	ListExperimentTemplates(ctx context.Context, params *fis.ListExperimentTemplatesInput,
		optFns ...func(*fis.Options)) (*fis.ListExperimentTemplatesOutput, error)
	DeleteExperimentTemplate(ctx context.Context, params *fis.DeleteExperimentTemplateInput,
		optFns ...func(*fis.Options)) (*fis.DeleteExperimentTemplateOutput, error)
	ListExperiments(ctx context.Context, params *fis.ListExperimentsInput,
		optFns ...func(*fis.Options)) (*fis.ListExperimentsOutput, error)
	StopExperiment(ctx context.Context, params *fis.StopExperimentInput,
		optFns ...func(*fis.Options)) (*fis.StopExperimentOutput, error)
	ListTargetAccountConfigurations(ctx context.Context, params *fis.ListTargetAccountConfigurationsInput,
		optFns ...func(*fis.Options)) (*fis.ListTargetAccountConfigurationsOutput, error)
	DeleteTargetAccountConfiguration(ctx context.Context, params *fis.DeleteTargetAccountConfigurationInput,
		optFns ...func(*fis.Options)) (*fis.DeleteTargetAccountConfigurationOutput, error)
}
