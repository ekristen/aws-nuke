package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lexmodelbuildingservice"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const LexSlotTypeResource = "LexSlotType"

func init() {
	registry.Register(&registry.Registration{
		Name:   LexSlotTypeResource,
		Scope:  nuke.Account,
		Lister: &LexSlotTypeLister{},
	})
}

type LexSlotTypeLister struct{}

func (l *LexSlotTypeLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := lexmodelbuildingservice.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &lexmodelbuildingservice.GetSlotTypesInput{
		MaxResults: aws.Int64(20),
	}

	for {
		output, err := svc.GetSlotTypes(params)
		if err != nil {
			return nil, err
		}

		for _, bot := range output.SlotTypes {
			resources = append(resources, &LexSlotType{
				svc:  svc,
				name: bot.Name,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type LexSlotType struct {
	svc  *lexmodelbuildingservice.LexModelBuildingService
	name *string
}

func (f *LexSlotType) Remove(_ context.Context) error {
	_, err := f.svc.DeleteSlotType(&lexmodelbuildingservice.DeleteSlotTypeInput{
		Name: f.name,
	})

	return err
}

func (f *LexSlotType) String() string {
	return *f.name
}

func (f *LexSlotType) Properties() types.Properties {
	properties := types.NewProperties()

	properties.
		Set("Name", f.name)
	return properties
}
