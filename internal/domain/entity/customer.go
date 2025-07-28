package entity

import (
	"time"

	"github.com/shopspring/decimal"
)

type Customer struct {
	CustomerID       string
	Points           int64
	LastPurchaseDate time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
	Records          []Record
	PointsByDate     map[string]int64
}

type Record struct {
	ProductID    string
	BranchID     string
	Amount       decimal.Decimal
	PurchaseDate time.Time
}
type UpdateCustomer struct {
	CustomerID       string
	PointsToAdd      int64
	LastPurchaseDate time.Time
	Records          []Record
	PointsByDate     map[string]int64
}

type Customers []Customer

func (customers Customers) GetCustomerByID(id string) (Customer, bool) {
	for _, customer := range customers {
		if customer.CustomerID == id {
			return customer, true
		}
	}
	return Customer{}, false
}
