package resources

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"            //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/signer" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const SignerSigningJobResource = "SignerSigningJob"

func init() {
	registry.Register(&registry.Registration{
		Name:     SignerSigningJobResource,
		Scope:    nuke.Account,
		Resource: &SignerSigningJob{},
		Lister:   &SignerSigningJobLister{},
	})
}

type SignerSigningJobLister struct{}

func (l *SignerSigningJobLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := signer.New(opts.Session)
	resources := make([]resource.Resource, 0)
	const reason string = "Revoked by AWS Nuke"

	listJobsInput := &signer.ListSigningJobsInput{}

	err := svc.ListSigningJobsPages(listJobsInput, func(page *signer.ListSigningJobsOutput, lastPage bool) bool {
		for _, job := range page.Jobs {
			resources = append(resources, &SignerSigningJob{
				svc:                 svc,
				jobID:               job.JobId,
				reason:              reason,
				isRevoked:           job.IsRevoked,
				createdAt:           *job.CreatedAt,
				profileName:         job.ProfileName,
				profileVersion:      job.ProfileVersion,
				platformID:          job.PlatformId,
				platformDisplayName: job.PlatformDisplayName,
				jobOwner:            job.JobOwner,
				jobInvoker:          job.JobInvoker,
			})
		}
		return true // continue iterating over pages
	})
	if err != nil {
		return nil, err
	}
	return resources, nil
}

type SignerSigningJob struct {
	svc                 *signer.Signer
	jobID               *string
	reason              string
	isRevoked           *bool
	createdAt           time.Time
	profileName         *string
	profileVersion      *string
	platformID          *string
	platformDisplayName *string
	jobOwner            *string
	jobInvoker          *string
}

func (j *SignerSigningJob) Filter() error {
	// Consider all non-revoked jobs
	if *j.isRevoked {
		return fmt.Errorf("job already revoked")
	}
	return nil
}

func (j *SignerSigningJob) Remove(_ context.Context) error {
	// Signing jobs are viewable by the ListSigningJobs operation for two years after they are performed [1]
	// As a precaution we are updating Signing jobs statuses to revoked. This indicates that the signature is no longer valid.
	// [1] https://awscli.amazonaws.com/v2/documentation/api/latest/reference/signer/start-signing-job.html
	revokeInput := &signer.RevokeSignatureInput{
		JobId:  j.jobID,
		Reason: aws.String(j.reason),
	}
	_, err := j.svc.RevokeSignature(revokeInput)
	return err
}

func (j *SignerSigningJob) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("JobId", j.jobID)
	properties.Set("CreatedAt", j.createdAt.Format(time.RFC3339))
	properties.Set("ProfileName", j.profileName)
	properties.Set("ProfileVersion", j.profileVersion)
	properties.Set("PlatformId", j.platformID)
	properties.Set("PlatformDisplayName", j.platformDisplayName)
	properties.Set("JobOwner", j.jobOwner)
	properties.Set("JobInvoker", j.jobInvoker)
	return properties
}
