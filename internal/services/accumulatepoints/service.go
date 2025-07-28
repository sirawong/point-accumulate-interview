package accumulatepoints

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"time"

	"github.com/gocarina/gocsv"
	"github.com/shopspring/decimal"
	"github.com/sirawong/point-accumulate-interview/internal/domain/entity"
	"github.com/sirawong/point-accumulate-interview/internal/domain/repository"
	apperr "github.com/sirawong/point-accumulate-interview/internal/errors"
	"github.com/sirawong/point-accumulate-interview/pkg/config"
	"github.com/sirawong/point-accumulate-interview/pkg/csv"
)

type accumulatePointService struct {
	ruleRepo     repository.RuleRepository
	customerRepo repository.CustomerRepository
	cfg          *config.Config
}

//go:generate mockgen -source=service.go -destination=mocks/mock_service.go -package=mocks
type AccumulatePointService interface {
	ExecuteMultipleFiles(ctx context.Context, files []FileInput) error
}

func NewAccumulatePointService(ruleRepo repository.RuleRepository, customerRepo repository.CustomerRepository, cfg *config.Config) AccumulatePointService {
	return &accumulatePointService{
		ruleRepo:     ruleRepo,
		customerRepo: customerRepo,
		cfg:          cfg,
	}
}

func (a accumulatePointService) ExecuteMultipleFiles(ctx context.Context, files []FileInput) error {
	if len(files) == 0 {
		return apperr.ErrInvalidArgument.WithMessage("no files")
	}

	allRecords := make(PurchaseRecords, 0)
	purchasedDate := make(map[time.Time]struct{})

	for _, file := range files {
		var records PurchaseRecords
		if err := csv.UnmarshalWithHeaderValidation(file.Reader, &records); err != nil {
			return apperr.ErrInvalidArgument.Wrap(err)
		}

		purchasedDate[file.PurchasedDate] = struct{}{}
		for _, record := range records {
			record.PurchaseDate = file.PurchasedDate
			allRecords = append(allRecords, record)
		}
	}

	allRecords = allRecords.getUniqueRecords()

	rules, err := a.ruleRepo.GetActiveRules(ctx, allRecords.mapRecordsToBranchCategories())
	if err != nil {
		return err
	}

	customers, err := a.customerRepo.GetCustomers(ctx, allRecords.getUniqueCustomerIDs())
	if err != nil {
		return err
	}

	customerAggregates := calculateBatchPoints(rules, allRecords, customers)
	if len(customerAggregates) > 0 {
		err = a.customerRepo.UpdateBulkCustomers(ctx, customerAggregates)
		if err != nil {
			return err
		}
	}

	customersUpdated, err := a.customerRepo.GetCustomers(ctx, []string{})
	if err != nil {
		return err
	}

	for date := range purchasedDate {
		dateString := date.Format(time.DateOnly)
		err = saveFile(customersUpdated, a.cfg.FilePath, dateString)
		if err != nil {
			return err
		}
	}

	return nil
}

func saveFile(customers []entity.Customer, filePath, dateString string) error {
	fullPath := fmt.Sprintf(filePath, dateString)

	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return apperr.ErrInternal.Wrap(fmt.Errorf("failed to create directory %s: %w", dir, err))
	}

	file, err := os.Create(fullPath)
	if err != nil {
		return apperr.ErrInternal.Wrap(err)
	}
	defer file.Close()

	records := entityToCustomerPointRecord(customers, dateString)

	sort.Slice(records, func(i, j int) bool {
		if records[i].Points != records[j].Points {
			return records[i].Points > records[j].Points
		}

		return records[i].LastPurchaseDate > records[j].LastPurchaseDate
	})

	if err = gocsv.MarshalFile(&records, file); err != nil {
		return apperr.ErrInternal.Wrap(err)
	}

	return nil
}

func calculateBatchPoints(rules []entity.Rule, records PurchaseRecords, customers entity.Customers) []entity.UpdateCustomer {
	recordsSetup := make(map[string]entity.UpdateCustomer, len(customers))

	for _, record := range records {
		for _, rule := range rules {
			points, applied := validateAndCalculatePoints(rule, *record, customers)
			if !applied {
				continue
			}

			customerID := record.CustomerID
			if _, ok := recordsSetup[customerID]; !ok {
				recordsSetup[customerID] = entity.UpdateCustomer{
					CustomerID:       customerID,
					PointsToAdd:      points,
					LastPurchaseDate: record.PurchaseDate,
					Records: []entity.Record{
						{
							ProductID:    record.ProductID,
							BranchID:     record.BranchID,
							Amount:       record.PurchasedAmount,
							PurchaseDate: record.PurchaseDate,
						},
					},
					PointsByDate: map[string]int64{record.PurchaseDate.Format(time.DateOnly): points},
				}
			} else {
				customer := recordsSetup[customerID]
				customer.PointsToAdd += points
				customer.Records = append(customer.Records, entity.Record{
					ProductID:    record.ProductID,
					BranchID:     record.BranchID,
					Amount:       record.PurchasedAmount,
					PurchaseDate: record.PurchaseDate,
				})

				if record.PurchaseDate.After(customer.LastPurchaseDate) {
					customer.LastPurchaseDate = record.PurchaseDate
				}
				customer.PointsByDate[record.PurchaseDate.Format(time.DateOnly)] += points

				recordsSetup[customerID] = customer
			}
		}
	}

	if len(recordsSetup) == 0 {
		return nil
	}

	result := make([]entity.UpdateCustomer, 0, len(recordsSetup))
	for _, customer := range recordsSetup {
		result = append(result, customer)
	}

	return result
}

func validateAndCalculatePoints(rule entity.Rule, record PurchaseRecord, customers entity.Customers) (points int64, applied bool) {
	customer, found := customers.GetCustomerByID(record.CustomerID)
	if found {
		isDuplicateRecord := slices.ContainsFunc(customer.Records, func(r entity.Record) bool {
			return r.PurchaseDate.Equal(record.PurchaseDate) &&
				r.BranchID == record.BranchID &&
				r.ProductID == record.ProductID &&
				r.Amount.Equal(record.PurchasedAmount)
		})

		if isDuplicateRecord {
			return 0, false
		}
	}

	if record.PurchasedAmount.LessThan(rule.Conditions.MinAmount) {
		return 0, false
	}

	if rule.Conditions.BranchID != "" && rule.Conditions.BranchID != record.BranchID {
		return 0, false
	}

	if len(rule.Conditions.CategoryID) > 0 && !slices.Contains(rule.Conditions.CategoryID, record.CategoryID) {
		return 0, false
	}

	switch rule.RuleType {
	case entity.FixedPointRule:
		return rule.Reward.Value, true

	case entity.PercentageRule:
		pointsDecimal := record.PurchasedAmount.Mul(decimal.NewFromInt(rule.Reward.Value)).Div(decimal.NewFromInt(100))
		return pointsDecimal.IntPart(), true

	case entity.RatioRule:
		if rule.Reward.RatioUnit != nil && *rule.Reward.RatioUnit > 0 {
			ratioUnitDecimal := decimal.NewFromFloat(*rule.Reward.RatioUnit)
			pointsDecimal := record.PurchasedAmount.Div(ratioUnitDecimal).Floor().Mul(decimal.NewFromInt(rule.Reward.Value))
			return pointsDecimal.IntPart(), true
		}
	}

	return 0, false
}
