package resources

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/budgets"
	"github.com/aws/aws-sdk-go/service/sts"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const BudgetResource = "Budget"

func init() {
	registry.Register(&registry.Registration{
		Name:   BudgetResource,
		Scope:  nuke.Account,
		Lister: &BudgetLister{},
	})
}

type BudgetLister struct{}

func (l *BudgetLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := budgets.New(opts.Session)

	// TODO: modify ListerOpts to include Account to reduce API calls
	identityOutput, err := sts.New(opts.Session).GetCallerIdentity(nil)
	if err != nil {
		fmt.Printf("sts error: %s \n", err)
		return nil, err
	}
	accountID := identityOutput.Account

	params := &budgets.DescribeBudgetsInput{
		AccountId:  aws.String(*accountID),
		MaxResults: aws.Int64(100),
	}

	buds := make([]*budgets.Budget, 0)
	err = svc.DescribeBudgetsPages(params, func(page *budgets.DescribeBudgetsOutput, lastPage bool) bool {
		buds = append(buds, page.Budgets...)
		return true
	})

	if err != nil {
		return nil, err
	}

	var resources []resource.Resource
	for _, bud := range buds {
		resources = append(resources, &Budget{
			svc:        svc,
			name:       bud.BudgetName,
			budgetType: bud.BudgetType,
			accountID:  accountID,
		})
	}

	return resources, nil
}

type Budget struct {
	svc        *budgets.Budgets
	name       *string
	budgetType *string
	accountID  *string
}

func (b *Budget) Remove(_ context.Context) error {
	_, err := b.svc.DeleteBudget(&budgets.DeleteBudgetInput{
		AccountId:  b.accountID,
		BudgetName: b.name,
	})

	return err
}

func (b *Budget) Properties() types.Properties {
	properties := types.NewProperties()

	properties.
		Set("Name", *b.name).
		Set("BudgetType", *b.budgetType).
		Set("AccountID", *b.accountID)
	return properties
}

func (b *Budget) String() string {
	return *b.name
}
