package resources

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"

	"github.com/aws/aws-sdk-go/service/budgets" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/v3/mocks/mock_budgetsiface"
)

func Test_Mock_BudgetsBudget_List(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBudgets := mock_budgetsiface.NewMockBudgetsAPI(ctrl)

	mockBudgets.EXPECT().DescribeBudgetsPages(gomock.Any(), gomock.Any()).DoAndReturn(
		func(input *budgets.DescribeBudgetsInput, fn func(*budgets.DescribeBudgetsOutput, bool) bool) error {
			fn(&budgets.DescribeBudgetsOutput{
				Budgets: []*budgets.Budget{
					{
						BudgetName: ptr.String("budget1"),
						BudgetType: ptr.String("COST"),
						BudgetLimit: &budgets.Spend{
							Amount: ptr.String("100"),
							Unit:   ptr.String("USD"),
						},
					},
				},
			}, false)
			fn(&budgets.DescribeBudgetsOutput{
				Budgets: []*budgets.Budget{
					{
						BudgetName: ptr.String("budget2"),
						BudgetType: ptr.String("COST"),
						BudgetLimit: &budgets.Spend{
							Amount: ptr.String("200"),
							Unit:   ptr.String("USD"),
						},
					},
				},
			}, true)
			return nil
		})

	mockBudgets.EXPECT().ListTagsForResource(&budgets.ListTagsForResourceInput{
		ResourceARN: ptr.String("arn:aws:budgets::012345678901:budget/budget1"),
	}).Return(&budgets.ListTagsForResourceOutput{
		ResourceTags: []*budgets.ResourceTag{
			{
				Key:   ptr.String("key1"),
				Value: ptr.String("value1"),
			},
		},
	}, nil)

	mockBudgets.EXPECT().ListTagsForResource(&budgets.ListTagsForResourceInput{
		ResourceARN: ptr.String("arn:aws:budgets::012345678901:budget/budget2"),
	}).Return(&budgets.ListTagsForResourceOutput{}, nil)

	lister := &BudgetsBudgetLister{
		mockSvc: mockBudgets,
	}
	resources, err := lister.List(context.TODO(), testListerOpts)
	a.Nil(err)
	a.Len(resources, 2)

	expectedResources := []resource.Resource{
		&BudgetsBudget{
			svc:        mockBudgets,
			Name:       ptr.String("budget1"),
			BudgetType: ptr.String("COST"),
			AccountID:  ptr.String("012345678901"),
			Tags: []*budgets.ResourceTag{
				{
					Key:   ptr.String("key1"),
					Value: ptr.String("value1"),
				},
			},
		},
		&BudgetsBudget{
			svc:        mockBudgets,
			Name:       ptr.String("budget2"),
			BudgetType: ptr.String("COST"),
			AccountID:  ptr.String("012345678901"),
		},
	}

	a.Equal(expectedResources, resources)
}

func Test_Mock_Budget_Remove(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBudgets := mock_budgetsiface.NewMockBudgetsAPI(ctrl)

	mockBudgets.EXPECT().DeleteBudget(gomock.Any()).Return(&budgets.DeleteBudgetOutput{}, nil)

	budget := &BudgetsBudget{
		svc:  mockBudgets,
		Name: ptr.String("budget1"),
	}

	err := budget.Remove(context.TODO())
	a.Nil(err)
}

func Test_Mock_Budget_Properties(t *testing.T) {
	a := assert.New(t)

	budget := &BudgetsBudget{
		Name:       ptr.String("budget1"),
		BudgetType: ptr.String("COST"),
		AccountID:  ptr.String("012345678901"),
	}

	properties := budget.Properties()
	a.Equal("budget1", properties.Get("Name"))
	a.Equal("COST", properties.Get("BudgetType"))
	a.Equal("012345678901", properties.Get("AccountID"))

	a.Equal("budget1", budget.String())
}
