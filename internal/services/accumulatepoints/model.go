package accumulatepoints

import (
	"io"
	"time"

	"github.com/shopspring/decimal"
	"github.com/sirawong/point-accumulate-interview/internal/domain/entity"
)

type PurchaseRecord struct {
	CustomerID      string          `csv:"customer_id"`
	ProductID       string          `csv:"product_id"`
	CategoryID      string          `csv:"category_id"`
	CategoryName    string          `csv:"category_name"`
	BranchID        string          `csv:"branch_id"`
	PurchasedAmount decimal.Decimal `csv:"purchased_amount"`
	Currency        string          `csv:"currency"`
	PurchaseDate    time.Time       `csv:"-"`
}

type FileInput struct {
	PurchasedDate time.Time
	Reader        io.ReadSeeker
}

type CustomerPointRecord struct {
	CustomerID       string `csv:"customer_id"`
	Points           int64  `csv:"points"`
	LastPurchaseDate string `csv:"last_purchase_date"`
}

func entityToCustomerPointRecord(customers []entity.Customer, targetDate string) []*CustomerPointRecord {
	result := make([]*CustomerPointRecord, 0, len(customers))

	for _, customer := range customers {
		points, lastDate := sumValuesOnOrBeforeDate(customer.PointsByDate, targetDate)
		if lastDate == "" {
			continue
		}

		result = append(result, &CustomerPointRecord{
			CustomerID:       customer.CustomerID,
			Points:           points,
			LastPurchaseDate: lastDate,
		})
	}

	return result
}

func sumValuesOnOrBeforeDate(data map[string]int64, targetDate string) (total int64, lastDate string) {
	for dateStr, value := range data {
		if dateStr <= targetDate {
			total += value
			if dateStr > lastDate {
				lastDate = dateStr
			}
		}
	}

	return total, lastDate
}

type PurchaseRecords []*PurchaseRecord

func (records PurchaseRecords) getUniqueCustomerIDs() []string {
	customerIDUnique := make(map[string]struct{})
	for _, record := range records {
		customerIDUnique[record.CustomerID] = struct{}{}
	}

	result := make([]string, 0, len(records))
	for id := range customerIDUnique {
		result = append(result, id)
	}

	return result
}

func (records PurchaseRecords) getUniqueRecords() PurchaseRecords {
	type recordKey struct {
		CustomerID      string
		ProductID       string
		BranchID        string
		PurchasedAmount string
		PurchaseDate    time.Time
	}

	seen := make(map[recordKey]struct{})
	uniqueRecords := make(PurchaseRecords, 0, len(records))

	for _, record := range records {
		key := recordKey{
			CustomerID:      record.CustomerID,
			ProductID:       record.ProductID,
			BranchID:        record.BranchID,
			PurchasedAmount: record.PurchasedAmount.String(),
			PurchaseDate:    record.PurchaseDate,
		}

		if _, ok := seen[key]; !ok {
			seen[key] = struct{}{}
			uniqueRecords = append(uniqueRecords, record)
		}
	}

	return uniqueRecords
}

func (records PurchaseRecords) mapRecordsToBranchCategories() map[string][]string {
	helperMap := make(map[string]map[string]struct{})

	for _, record := range records {
		if _, ok := helperMap[record.BranchID]; !ok {
			helperMap[record.BranchID] = make(map[string]struct{})
		}
		helperMap[record.BranchID][record.CategoryID] = struct{}{}
	}

	result := make(map[string][]string)
	for branchID, categorySet := range helperMap {
		categories := make([]string, 0, len(categorySet))
		for categoryID := range categorySet {
			categories = append(categories, categoryID)
		}
		result[branchID] = categories
	}

	return result
}
