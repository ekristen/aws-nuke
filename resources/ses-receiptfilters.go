package resources

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ses"

	sdkerrors "github.com/ekristen/libnuke/pkg/errors"
	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/v3/pkg/awsutil"
	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const SESReceiptFilterResource = "SESReceiptFilter"

func init() {
	registry.Register(&registry.Registration{
		Name:   SESReceiptFilterResource,
		Scope:  nuke.Account,
		Lister: &SESReceiptFilterLister{},
	})
}

type SESReceiptFilterLister struct{}

func (l *SESReceiptFilterLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := ses.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &ses.ListReceiptFiltersInput{}

	output, err := svc.ListReceiptFilters(params)
	if err != nil {
		// SES capabilities aren't the same in all regions, for example us-west-1 will throw InvalidAction
		// errors, but other regions work, this allows us to safely ignore these and yet log them in debug logs
		// should we need to troubleshoot.
		var awsError awserr.Error
		if errors.As(err, &awsError) {
			if awsError.Code() == awsutil.ErrCodeInvalidAction {
				return nil, sdkerrors.ErrSkipRequest(
					"Listing of SESReceiptFilter not supported in this region: " + *opts.Session.Config.Region)
			}
		}

		return nil, err
	}

	for _, filter := range output.Filters {
		resources = append(resources, &SESReceiptFilter{
			svc:  svc,
			name: filter.Name,
		})
	}

	return resources, nil
}

type SESReceiptFilter struct {
	svc  *ses.SES
	name *string
}

func (f *SESReceiptFilter) Remove(_ context.Context) error {
	_, err := f.svc.DeleteReceiptFilter(&ses.DeleteReceiptFilterInput{
		FilterName: f.name,
	})

	return err
}

func (f *SESReceiptFilter) String() string {
	return *f.name
}
