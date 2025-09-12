package resources

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/mgn"
	"github.com/aws/aws-sdk-go-v2/service/mgn/types"
	"github.com/aws/smithy-go"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	libtypes "github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const (
	MGNWaveResource                      = "MGNWave"
	mgnWaveUninitializedAccountException = "UninitializedAccountException"
)

func init() {
	registry.Register(&registry.Registration{
		Name:     MGNWaveResource,
		Scope:    nuke.Account,
		Resource: &MGNWave{},
		Lister:   &MGNWaveLister{},
	})
}

type MGNWaveLister struct{}

func (l *MGNWaveLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := mgn.NewFromConfig(*opts.Config)
	resources := make([]resource.Resource, 0)

	params := &mgn.ListWavesInput{
		MaxResults: aws.Int32(50),
	}

	for {
		output, err := svc.ListWaves(ctx, params)
		if err != nil {
			var apiErr smithy.APIError
			if errors.As(err, &apiErr) && apiErr.ErrorCode() == mgnWaveUninitializedAccountException {
				return nil, nil
			}
			return nil, err
		}

		for i := range output.Items {
			wave := &output.Items[i]
			mgnWave := &MGNWave{
				svc:                  svc,
				wave:                 wave,
				WaveID:               wave.WaveID,
				Arn:                  wave.Arn,
				Name:                 wave.Name,
				Description:          wave.Description,
				IsArchived:           wave.IsArchived,
				CreationDateTime:     wave.CreationDateTime,
				LastModifiedDateTime: wave.LastModifiedDateTime,
				Tags:                 wave.Tags,
			}
			resources = append(resources, mgnWave)
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type MGNWave struct {
	svc  *mgn.Client `description:"-"`
	wave *types.Wave `description:"-"`

	// Exposed properties
	WaveID               *string           `description:"The unique identifier of the wave"`
	Arn                  *string           `description:"The ARN of the wave"`
	Name                 *string           `description:"The name of the wave"`
	Description          *string           `description:"The description of the wave"`
	IsArchived           *bool             `description:"Whether the wave is archived"`
	CreationDateTime     *string           `description:"The date and time the wave was created"`
	LastModifiedDateTime *string           `description:"The date and time the wave was last modified"`
	Tags                 map[string]string `description:"The tags associated with the wave"`
}

func (f *MGNWave) Remove(ctx context.Context) error {
	_, err := f.svc.DeleteWave(ctx, &mgn.DeleteWaveInput{
		WaveID: f.wave.WaveID,
	})

	return err
}

func (f *MGNWave) Properties() libtypes.Properties {
	return libtypes.NewPropertiesFromStruct(f)
}

func (f *MGNWave) String() string {
	return *f.WaveID
}
