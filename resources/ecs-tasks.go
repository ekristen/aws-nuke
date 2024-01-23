package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"

	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const ECSTaskResource = "ECSTask"

func init() {
	resource.Register(&resource.Registration{
		Name:   ECSTaskResource,
		Scope:  nuke.Account,
		Lister: &ECSTaskLister{},
	})
}

type ECSTaskLister struct{}

func (l *ECSTaskLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := ecs.New(opts.Session)
	resources := make([]resource.Resource, 0)
	var clusters []*string

	clusterParams := &ecs.ListClustersInput{
		MaxResults: aws.Int64(100),
	}

	// Discover all clusters
	for {
		output, err := svc.ListClusters(clusterParams)
		if err != nil {
			return nil, err
		}

		clusters = append(clusters, output.ClusterArns...)

		if output.NextToken == nil {
			break
		}

		clusterParams.NextToken = output.NextToken
	}

	// Discover all running tasks from all clusters
	for _, clusterArn := range clusters {
		taskParams := &ecs.ListTasksInput{
			Cluster:       clusterArn,
			MaxResults:    aws.Int64(10),
			DesiredStatus: aws.String("RUNNING"),
		}
		output, err := svc.ListTasks(taskParams)
		if err != nil {
			return nil, err
		}

		for _, taskArn := range output.TaskArns {
			resources = append(resources, &ECSTask{
				svc:        svc,
				taskARN:    taskArn,
				clusterARN: clusterArn,
			})
		}

		if output.NextToken == nil {
			continue
		}

		taskParams.NextToken = output.NextToken
	}

	return resources, nil
}

type ECSTask struct {
	svc        *ecs.ECS
	taskARN    *string
	clusterARN *string
}

func (t *ECSTask) Filter() error {
	return nil
}

func (t *ECSTask) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("TaskARN", t.taskARN)
	properties.Set("ClusterARN", t.clusterARN)

	return properties
}

func (t *ECSTask) Remove(_ context.Context) error {
	// When StopTask is called on a task, the equivalent of docker stop is issued to the
	// containers running in the task. This results in a SIGTERM value and a default
	// 30-second timeout, after which the SIGKILL value is sent and the containers are
	// forcibly stopped. If the container handles the SIGTERM value gracefully and exits
	// within 30 seconds from receiving it, no SIGKILL value is sent.
	_, err := t.svc.StopTask(&ecs.StopTaskInput{
		Cluster: t.clusterARN,
		Task:    t.taskARN,
		Reason:  aws.String("Task stopped via AWS Nuke"),
	})

	return err
}
