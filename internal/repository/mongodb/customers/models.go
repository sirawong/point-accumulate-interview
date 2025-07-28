package mongodb

import (
	"log"
	"time"

	"github.com/shopspring/decimal"
	"github.com/sirawong/point-accumulate-interview/internal/errors"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/sirawong/point-accumulate-interview/internal/domain/entity"
)

type Customer struct {
	ID               primitive.ObjectID `bson:"_id,omitempty"`
	CustomerID       string             `bson:"customer_id"`
	Points           int64              `bson:"points"`
	LastPurchaseDate time.Time          `bson:"last_purchase_date"`
	CreatedAt        time.Time          `bson:"created_at"`
	UpdatedAt        time.Time          `bson:"updated_at"`
	Records          []Record           `bson:"records"`
	PointsByDate     map[string]int64   `bson:"points_by_date"`
}

type Record struct {
	ProductID    string               `bson:"product_id"`
	BranchID     string               `bson:"branch_id"`
	Amount       primitive.Decimal128 `bson:"amount"`
	PurchaseDate time.Time            `bson:"purchase_date"`
}

func (u Customer) ToDomain() (*entity.Customer, error) {
	records := make([]entity.Record, 0, len(u.Records))
	for _, record := range u.Records {

		value, err := decimal.NewFromString(record.Amount.String())
		if err != nil {
			return nil, errors.ErrInternal.Wrap(err)
		}

		records = append(records, entity.Record{
			ProductID:    record.ProductID,
			BranchID:     record.BranchID,
			Amount:       value,
			PurchaseDate: record.PurchaseDate,
		})
	}

	return &entity.Customer{
		CustomerID:       u.CustomerID,
		Points:           u.Points,
		LastPurchaseDate: u.LastPurchaseDate,
		CreatedAt:        u.CreatedAt,
		UpdatedAt:        u.UpdatedAt,
		Records:          records,
		PointsByDate:     u.PointsByDate,
	}, nil
}

type Customers []*Customer

func (u Customers) ToDomain() []entity.Customer {
	result := make([]entity.Customer, 0, len(u))
	for _, customer := range u {
		value, err := customer.ToDomain()
		if err != nil {
			log.Println(errors.ErrInternal, err)
			continue
		}
		result = append(result, *value)
	}

	return result
}

func fromRecords(records []entity.Record) ([]Record, error) {
	result := make([]Record, 0, len(records))
	for _, record := range records {
		value, err := primitive.ParseDecimal128(record.Amount.String())
		if err != nil {
			return nil, errors.ErrInvalidArgument.Wrap(err)
		}

		result = append(result, Record{
			ProductID:    record.ProductID,
			BranchID:     record.BranchID,
			Amount:       value,
			PurchaseDate: record.PurchaseDate,
		})
	}
	return result, nil
}
