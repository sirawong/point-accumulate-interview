package mongodb

import (
	"context"

	"github.com/sirawong/point-accumulate-interview/internal/domain/entity"
	"github.com/sirawong/point-accumulate-interview/internal/domain/repository"
	"github.com/sirawong/point-accumulate-interview/internal/errors"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	CustomerCollection = "customers"
)

type customerRepository struct {
	collection *mongo.Collection
}

func NewCustomerRepository(db *mongo.Database) repository.CustomerRepository {
	return &customerRepository{
		collection: db.Collection(CustomerCollection),
	}
}

func (c customerRepository) GetCustomers(ctx context.Context, customerIDs []string) ([]entity.Customer, error) {
	cursor, err := c.collection.Find(ctx, filterCustomerIDs(customerIDs), nil)
	if err != nil {
		return nil, errors.ErrInternal.Wrap(err)
	}

	var customers Customers
	err = cursor.All(ctx, &customers)
	if err != nil {
		return nil, errors.ErrInternal.Wrap(err)
	}

	return customers.ToDomain(), nil
}

func (c customerRepository) UpdateBulkCustomers(ctx context.Context, updateCustomer []entity.UpdateCustomer) error {
	operations, err := operationUpdateCustomerPoint(updateCustomer)
	if err != nil {
		return err
	}

	opts := options.BulkWrite().SetOrdered(false)
	_, err = c.collection.BulkWrite(ctx, operations, opts)
	if err != nil {
		return errors.ErrInternal.Wrap(err)
	}

	return nil
}
