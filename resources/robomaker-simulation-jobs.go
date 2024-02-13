package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/robomaker"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const RoboMakerSimulationJobResource = "RoboMakerSimulationJob"

func init() {
	registry.Register(&registry.Registration{
		Name:   RoboMakerSimulationJobResource,
		Scope:  nuke.Account,
		Lister: &RoboMakerSimulationJobLister{},
	})
}

type RoboMakerSimulationJobLister struct{}

func (l *RoboMakerSimulationJobLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := robomaker.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &robomaker.ListSimulationJobsInput{
		MaxResults: aws.Int64(30),
	}

	for {
		resp, err := svc.ListSimulationJobs(params)
		if err != nil {
			return nil, err
		}

		for _, simulationJob := range resp.SimulationJobSummaries {
			if simulationJobNeedsToBeCanceled(simulationJob) {
				resources = append(resources, &RoboMakerSimulationJob{
					svc: svc,
					arn: simulationJob.Arn,
				})
			}
		}

		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return resources, nil
}

// simulationJobNeedsToBeCanceled returns true if the simulation job needs to be canceled (helper function)
func simulationJobNeedsToBeCanceled(job *robomaker.SimulationJobSummary) bool {
	for _, n := range []string{"Completed", "Failed", "RunningFailed", "Terminating", "Terminated", "Canceled"} {
		if job.Status != nil && *job.Status == n {
			return false
		}
	}
	return true
}

type RoboMakerSimulationJob struct {
	svc  *robomaker.RoboMaker
	name *string
	arn  *string
}

func (f *RoboMakerSimulationJob) Remove(_ context.Context) error {
	_, err := f.svc.CancelSimulationJob(&robomaker.CancelSimulationJobInput{
		Job: f.arn,
	})

	return err
}

func (f *RoboMakerSimulationJob) String() string {
	return *f.arn
}
