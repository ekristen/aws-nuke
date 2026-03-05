package resources

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/translate"
)

type TranslateAPI interface {
	ListParallelData(ctx context.Context, params *translate.ListParallelDataInput,
		optFns ...func(*translate.Options)) (*translate.ListParallelDataOutput, error)
	DeleteParallelData(ctx context.Context, params *translate.DeleteParallelDataInput,
		optFns ...func(*translate.Options)) (*translate.DeleteParallelDataOutput, error)
	ListTerminologies(ctx context.Context, params *translate.ListTerminologiesInput,
		optFns ...func(*translate.Options)) (*translate.ListTerminologiesOutput, error)
	DeleteTerminology(ctx context.Context, params *translate.DeleteTerminologyInput,
		optFns ...func(*translate.Options)) (*translate.DeleteTerminologyOutput, error)
	ListTextTranslationJobs(ctx context.Context, params *translate.ListTextTranslationJobsInput,
		optFns ...func(*translate.Options)) (*translate.ListTextTranslationJobsOutput, error)
	StopTextTranslationJob(ctx context.Context, params *translate.StopTextTranslationJobInput,
		optFns ...func(*translate.Options)) (*translate.StopTextTranslationJobOutput, error)
}
