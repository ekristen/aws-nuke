package resources

import (
	"context"
	"github.com/ekristen/libnuke/pkg/settings"

	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/qldb"

	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const QLDBLedgerResource = "QLDBLedger"

func init() {
	resource.Register(&resource.Registration{
		Name:   QLDBLedgerResource,
		Scope:  nuke.Account,
		Lister: &QLDBLedgerLister{},
	})
}

type QLDBLedgerLister struct{}

func (l *QLDBLedgerLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := qldb.New(opts.Session)

	params := &qldb.ListLedgersInput{}
	resources := make([]resource.Resource, 0)

	for {
		resp, err := svc.ListLedgers(params)
		if err != nil {
			return nil, err
		}

		for _, ledger := range resp.Ledgers {
			ledgerDescription, err := svc.DescribeLedger(&qldb.DescribeLedgerInput{Name: ledger.Name})
			if err != nil {
				return nil, err
			}

			resources = append(resources, &QLDBLedger{
				svc:    svc,
				ledger: ledgerDescription,
			})
		}
		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return resources, nil
}

type QLDBLedger struct {
	svc    *qldb.QLDB
	ledger *qldb.DescribeLedgerOutput

	settings *settings.Setting
}

func (l *QLDBLedger) Settings(setting *settings.Setting) {
	l.settings = setting
}

func (l *QLDBLedger) Remove(_ context.Context) error {
	if aws.BoolValue(l.ledger.DeletionProtection) && l.settings.Get("DisableDeletionProtection").(bool) {
		modifyParams := &qldb.UpdateLedgerInput{
			DeletionProtection: aws.Bool(false),
			Name:               l.ledger.Name,
		}
		_, err := l.svc.UpdateLedger(modifyParams)
		if err != nil {
			return err
		}
	}

	params := &qldb.DeleteLedgerInput{
		Name: l.ledger.Name,
	}

	if _, err := l.svc.DeleteLedger(params); err != nil {
		return err
	}

	return nil
}

func (l *QLDBLedger) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("Name", l.ledger.Name)
	properties.Set("DeletionProtection", l.ledger.DeletionProtection)
	properties.Set("Arn", l.ledger.Arn)
	properties.Set("CreationDateTime", l.ledger.CreationDateTime.Format(time.RFC3339))
	properties.Set("State", l.ledger.State)
	properties.Set("PermissionsMode", l.ledger.PermissionsMode)
	properties.Set("EncryptionDescription", l.ledger.EncryptionDescription)
	return properties
}
func (l *QLDBLedger) String() string {
	return aws.StringValue(l.ledger.Name)
}
