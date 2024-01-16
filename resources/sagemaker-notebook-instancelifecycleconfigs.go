package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"

	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const SageMakerNotebookInstanceLifecycleConfigResource = "SageMakerNotebookInstanceLifecycleConfig"

func init() {
	resource.Register(resource.Registration{
		Name:   SageMakerNotebookInstanceLifecycleConfigResource,
		Scope:  nuke.Account,
		Lister: &SageMakerNotebookInstanceLifecycleConfigLister{},
	})
}

type SageMakerNotebookInstanceLifecycleConfigLister struct{}

func (l *SageMakerNotebookInstanceLifecycleConfigLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := sagemaker.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &sagemaker.ListNotebookInstanceLifecycleConfigsInput{
		MaxResults: aws.Int64(30),
	}

	for {
		resp, err := svc.ListNotebookInstanceLifecycleConfigs(params)
		if err != nil {
			return nil, err
		}

		for _, lifecycleConfig := range resp.NotebookInstanceLifecycleConfigs {
			resources = append(resources, &SageMakerNotebookInstanceLifecycleConfig{
				svc:  svc,
				Name: lifecycleConfig.NotebookInstanceLifecycleConfigName,
			})
		}

		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return resources, nil
}

type SageMakerNotebookInstanceLifecycleConfig struct {
	svc  *sagemaker.SageMaker
	Name *string
}

func (f *SageMakerNotebookInstanceLifecycleConfig) Remove(_ context.Context) error {

	_, err := f.svc.DeleteNotebookInstanceLifecycleConfig(&sagemaker.DeleteNotebookInstanceLifecycleConfigInput{
		NotebookInstanceLifecycleConfigName: f.Name,
	})

	return err
}

func (f *SageMakerNotebookInstanceLifecycleConfig) String() string {
	return *f.Name
}

func (f *SageMakerNotebookInstanceLifecycleConfig) Properties() types.Properties {
	properties := types.NewProperties()
	properties.
		Set("Name", f.Name)
	return properties
}
