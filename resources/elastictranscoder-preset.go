package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/service/elastictranscoder"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const ElasticTranscoderPresetResource = "ElasticTranscoderPreset"

func init() {
	registry.Register(&registry.Registration{
		Name:   ElasticTranscoderPresetResource,
		Scope:  nuke.Account,
		Lister: &ElasticTranscoderPresetLister{},
	})
}

type ElasticTranscoderPresetLister struct{}

func (l *ElasticTranscoderPresetLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := elastictranscoder.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &elastictranscoder.ListPresetsInput{}

	for {
		resp, err := svc.ListPresets(params)
		if err != nil {
			return nil, err
		}

		for _, preset := range resp.Presets {
			resources = append(resources, &ElasticTranscoderPreset{
				svc:      svc,
				PresetID: preset.Id,
			})
		}

		if resp.NextPageToken == nil {
			break
		}

		params.PageToken = resp.NextPageToken
	}

	return resources, nil
}

type ElasticTranscoderPreset struct {
	svc      *elastictranscoder.ElasticTranscoder
	PresetID *string
}

func (r *ElasticTranscoderPreset) Filter() error {
	if strings.HasPrefix(*r.PresetID, "1351620000001") {
		return fmt.Errorf("cannot delete elastic transcoder system presets")
	}
	return nil
}

func (r *ElasticTranscoderPreset) Remove(_ context.Context) error {
	_, err := r.svc.DeletePreset(&elastictranscoder.DeletePresetInput{
		Id: r.PresetID,
	})

	return err
}

func (r *ElasticTranscoderPreset) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *ElasticTranscoderPreset) String() string {
	return *r.PresetID
}
