package resources

import (
	"context"

	"github.com/gotidy/ptr"

	"github.com/aws/aws-sdk-go/service/budgets"
	"github.com/aws/aws-sdk-go/service/budgets/budgetsiface"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/aws/aws-sdk-go/service/sts/stsiface"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const BudgetsBudgetResource = "BudgetsBudget"

func init() {
	registry.Register(&registry.Registration{
		Name:   BudgetsBudgetResource,
		Scope:  nuke.Account,
		Lister: &BudgetsBudgetLister{},
		DeprecatedAliases: []string{
			"Budget",
		},
	})
}

type BudgetsBudgetLister struct {
	mockSvc    budgetsiface.BudgetsAPI
	mockSTSSvc stsiface.STSAPI
}

func (l *BudgetsBudgetLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	var resources []resource.Resource
	var svc budgetsiface.BudgetsAPI
	var stsSvc stsiface.STSAPI

	if l.mockSvc != nil {
		svc = l.mockSvc
	} else {
		svc = budgets.New(opts.Session)
	}

	if l.mockSTSSvc != nil {
		stsSvc = l.mockSTSSvc
	} else {
		stsSvc = sts.New(opts.Session)
	}

	// TODO: modify ListerOpts to include Account to reduce API calls
	identityOutput, err := stsSvc.GetCallerIdentity(nil)
	if err != nil {
		return nil, err
	}
	accountID := identityOutput.Account

	params := &budgets.DescribeBudgetsInput{
		AccountId:  accountID,
		MaxResults: ptr.Int64(100),
	}

	buds := make([]*budgets.Budget, 0)
	err = svc.DescribeBudgetsPages(params, func(page *budgets.DescribeBudgetsOutput, lastPage bool) bool {
		buds = append(buds, page.Budgets...)
		return true
	})

	if err != nil {
		return nil, err
	}

	for _, bud := range buds {
		resources = append(resources, &BudgetsBudget{
			svc:        svc,
			Name:       bud.BudgetName,
			BudgetType: bud.BudgetType,
			AccountID:  accountID,
		})
	}

	return resources, nil
}

type BudgetsBudget struct {
	svc        budgetsiface.BudgetsAPI
	Name       *string
	BudgetType *string
	AccountID  *string
}

func (r *BudgetsBudget) Remove(_ context.Context) error {
	_, err := r.svc.DeleteBudget(&budgets.DeleteBudgetInput{
		AccountId:  r.AccountID,
		BudgetName: r.Name,
	})

	return err
}

func (r *BudgetsBudget) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *BudgetsBudget) String() string {
	return *r.Name
}
