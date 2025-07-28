package repository

import (
	"context"

	"github.com/sirawong/point-accumulate-interview/internal/domain/entity"
)

//go:generate mockgen -source=repository.go -destination=mocks_rule/mock_repository.go -package=mocks_rule
type RuleRepository interface {
	GetActiveRules(ctx context.Context, branchIDWithCategoryIDs map[string][]string) ([]entity.Rule, error)
}

//go:generate mockgen -source=repository.go -destination=mocks_customer/mock_repository.go -package=mocks_customer
type CustomerRepository interface {
	GetCustomers(ctx context.Context, customerIDs []string) ([]entity.Customer, error)
	UpdateBulkCustomers(ctx context.Context, updateCustomer []entity.UpdateCustomer) error
}
