package accumulatepoints

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/sirawong/point-accumulate-interview/internal/domain/entity"
	"github.com/sirawong/point-accumulate-interview/internal/domain/repository/mocks_customer"
	"github.com/sirawong/point-accumulate-interview/internal/domain/repository/mocks_rule"
	apperr "github.com/sirawong/point-accumulate-interview/internal/errors"
	"github.com/sirawong/point-accumulate-interview/pkg/config"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

type AccumulatePointServiceTestSuite struct {
	suite.Suite
	mockCtrl     *gomock.Controller
	mockRuleRepo *mocks_rule.MockRuleRepository
	mockCustRepo *mocks_customer.MockCustomerRepository
	service      AccumulatePointService
	cfg          *config.Config
	tempDir      string
}

func TestAccumulatePointServiceTestSuite(t *testing.T) {
	suite.Run(t, new(AccumulatePointServiceTestSuite))
}

func (suite *AccumulatePointServiceTestSuite) SetupTest() {
	suite.mockCtrl = gomock.NewController(suite.T())
	suite.mockRuleRepo = mocks_rule.NewMockRuleRepository(suite.mockCtrl)
	suite.mockCustRepo = mocks_customer.NewMockCustomerRepository(suite.mockCtrl)

	tempDir, err := os.MkdirTemp("", "test_output_")
	suite.NoError(err)
	suite.tempDir = tempDir

	suite.cfg = &config.Config{
		FilePath: filepath.Join(tempDir, "output_%s.csv"),
	}
	suite.service = NewAccumulatePointService(suite.mockRuleRepo, suite.mockCustRepo, suite.cfg)
}

func (suite *AccumulatePointServiceTestSuite) TearDownTest() {
	suite.mockCtrl.Finish()

	if suite.tempDir != "" {
		err := os.RemoveAll(suite.tempDir)
		suite.NoError(err)
	}
}

func (suite *AccumulatePointServiceTestSuite) TestExecuteMultipleFiles_NoFiles() {
	err := suite.service.ExecuteMultipleFiles(context.Background(), []FileInput{})

	suite.Error(err)
	suite.Contains(err.Error(), "no files")
}

func (suite *AccumulatePointServiceTestSuite) TestExecuteMultipleFiles_CSVUnmarshalError() {
	ctx := context.Background()

	invalidCSV := "invalid,csv,data\nwithout,proper,headers"

	files := []FileInput{
		{
			PurchasedDate: time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
			Reader:        strings.NewReader(invalidCSV),
		},
	}

	err := suite.service.ExecuteMultipleFiles(ctx, files)

	suite.Error(err)
	suite.IsType(&apperr.AppError{}, err)
}

func (suite *AccumulatePointServiceTestSuite) TestExecuteMultipleFiles_NoPointsCalculated() {
	ctx := context.Background()
	purchaseDate := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)

	csvData := "customer_id,product_id,category_id,category_name,branch_id,purchased_amount,currency\n" +
		"U000001,123121,CT1001,ELECTRONICS,BR0001,10.00,THB"

	files := []FileInput{
		{
			PurchasedDate: purchaseDate,
			Reader:        strings.NewReader(csvData),
		},
	}

	rules := []entity.Rule{
		{
			ID:       "RULE001",
			RuleType: entity.FixedPointRule,
			Conditions: entity.Conditions{
				MinAmount: decimal.NewFromFloat(100.0),
			},
			Reward: entity.Reward{Value: 10},
		},
	}

	customers := []entity.Customer{
		{
			CustomerID:   "U000001",
			Points:       0,
			Records:      []entity.Record{},
			PointsByDate: make(map[string]int64),
		},
	}

	updatedCustomers := customers

	suite.mockRuleRepo.EXPECT().
		GetActiveRules(ctx, gomock.Any()).
		Return(rules, nil)

	suite.mockCustRepo.EXPECT().
		GetCustomers(ctx, []string{"U000001"}).
		Return(customers, nil)

	suite.mockCustRepo.EXPECT().
		GetCustomers(ctx, []string{}).
		Return(updatedCustomers, nil)

	err := suite.service.ExecuteMultipleFiles(ctx, files)

	suite.NoError(err)
}

func (suite *AccumulatePointServiceTestSuite) TestExecuteMultipleFiles_MultipleFiles() {
	ctx := context.Background()

	purchaseDate1 := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
	purchaseDate2 := time.Date(2025, 1, 16, 0, 0, 0, 0, time.UTC)

	csvData1 := "customer_id,product_id,category_id,category_name,branch_id,purchased_amount,currency\n" +
		"U000001,123121,CT1001,ELECTRONICS,BR0001,100.50,THB"

	csvData2 := "customer_id,product_id,category_id,category_name,branch_id,purchased_amount,currency\n" +
		"U000002,123122,CT1002,FASHION,BR0002,200.00,THB"

	files := []FileInput{
		{
			PurchasedDate: purchaseDate1,
			Reader:        strings.NewReader(csvData1),
		},
		{
			PurchasedDate: purchaseDate2,
			Reader:        strings.NewReader(csvData2),
		},
	}

	rules := []entity.Rule{
		{
			ID:       "RULE001",
			RuleType: entity.FixedPointRule,
			Conditions: entity.Conditions{
				MinAmount: decimal.NewFromFloat(50.0),
			},
			Reward: entity.Reward{Value: 10},
		},
	}

	customers := []entity.Customer{
		{
			CustomerID:   "U000001",
			Points:       0,
			Records:      []entity.Record{},
			PointsByDate: make(map[string]int64),
		},
		{
			CustomerID:   "U000002",
			Points:       0,
			Records:      []entity.Record{},
			PointsByDate: make(map[string]int64),
		},
	}

	updatedCustomers := []entity.Customer{
		{
			CustomerID: "U000001",
			Points:     10,
			PointsByDate: map[string]int64{
				"2025-01-15": 10,
			},
		},
		{
			CustomerID: "U000002",
			Points:     10,
			PointsByDate: map[string]int64{
				"2025-01-16": 10,
			},
		},
	}

	suite.mockRuleRepo.EXPECT().
		GetActiveRules(ctx, gomock.Any()).
		Return(rules, nil)

	suite.mockCustRepo.EXPECT().
		GetCustomers(ctx, gomock.Any()).
		Return(customers, nil)

	suite.mockCustRepo.EXPECT().
		UpdateBulkCustomers(ctx, gomock.Any()).
		Return(nil)

	suite.mockCustRepo.EXPECT().
		GetCustomers(ctx, []string{}).
		Return(updatedCustomers, nil)

	err := suite.service.ExecuteMultipleFiles(ctx, files)

	suite.NoError(err)

	expectedFiles := []string{
		fmt.Sprintf(suite.cfg.FilePath, "2025-01-15"),
		fmt.Sprintf(suite.cfg.FilePath, "2025-01-16"),
	}

	for _, expectedFile := range expectedFiles {
		suite.FileExists(expectedFile)
	}
}

func (suite *AccumulatePointServiceTestSuite) TestExecuteMultipleFiles_DuplicateRecords() {
	ctx := context.Background()
	purchaseDate := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)

	csvData := "customer_id,product_id,category_id,category_name,branch_id,purchased_amount,currency\n" +
		"U000001,123121,CT1001,ELECTRONICS,BR0001,100.50,THB\n" +
		"U000001,123121,CT1001,ELECTRONICS,BR0001,100.50,THB\n" +
		"U000002,123122,CT1002,FASHION,BR0002,200.00,THB"

	files := []FileInput{
		{
			PurchasedDate: purchaseDate,
			Reader:        strings.NewReader(csvData),
		},
	}

	rules := []entity.Rule{
		{
			ID:       "RULE001",
			RuleType: entity.FixedPointRule,
			Conditions: entity.Conditions{
				MinAmount: decimal.NewFromFloat(50.0),
			},
			Reward: entity.Reward{Value: 10},
		},
	}

	customers := []entity.Customer{
		{
			CustomerID:   "U000001",
			Points:       0,
			Records:      []entity.Record{},
			PointsByDate: make(map[string]int64),
		},
		{
			CustomerID:   "U000002",
			Points:       0,
			Records:      []entity.Record{},
			PointsByDate: make(map[string]int64),
		},
	}

	updatedCustomers := customers

	suite.mockRuleRepo.EXPECT().
		GetActiveRules(ctx, gomock.Any()).
		Return(rules, nil)

	suite.mockCustRepo.EXPECT().
		GetCustomers(ctx, gomock.Any()).
		DoAndReturn(func(ctx context.Context, customerIDs []string) ([]entity.Customer, error) {

			suite.Len(customerIDs, 2)
			suite.Contains(customerIDs, "U000001")
			suite.Contains(customerIDs, "U000002")
			return customers, nil
		})

	suite.mockCustRepo.EXPECT().
		UpdateBulkCustomers(ctx, gomock.Any()).
		DoAndReturn(func(ctx context.Context, updates []entity.UpdateCustomer) error {

			suite.Len(updates, 2)
			return nil
		})

	suite.mockCustRepo.EXPECT().
		GetCustomers(ctx, []string{}).
		Return(updatedCustomers, nil)

	err := suite.service.ExecuteMultipleFiles(ctx, files)

	suite.NoError(err)
}

func (suite *AccumulatePointServiceTestSuite) TestExecuteMultipleFiles_ComplexRulesScenario() {
	ctx := context.Background()
	purchaseDate := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)

	csvData := "customer_id,product_id,category_id,category_name,branch_id,purchased_amount,currency\n" +
		"U000001,123121,CT1001,ELECTRONICS,BR0001,100.00,THB\n" +
		"U000001,123122,CT1002,FASHION,BR0002,200.00,THB\n" +
		"U000002,123123,CT1001,ELECTRONICS,BR0001,150.00,THB"

	files := []FileInput{
		{
			PurchasedDate: purchaseDate,
			Reader:        strings.NewReader(csvData),
		},
	}

	rules := []entity.Rule{
		{
			ID:       "RULE001",
			RuleType: entity.FixedPointRule,
			Conditions: entity.Conditions{
				MinAmount:  decimal.NewFromFloat(50.0),
				BranchID:   "BR0001",
				CategoryID: []string{"CT1001"},
			},
			Reward: entity.Reward{Value: 20},
		},
		{
			ID:       "RULE002",
			RuleType: entity.PercentageRule,
			Conditions: entity.Conditions{
				MinAmount: decimal.NewFromFloat(100.0),
			},
			Reward: entity.Reward{Value: 5},
		},
	}

	customers := []entity.Customer{
		{
			CustomerID:   "U000001",
			Points:       0,
			Records:      []entity.Record{},
			PointsByDate: make(map[string]int64),
		},
		{
			CustomerID:   "U000002",
			Points:       0,
			Records:      []entity.Record{},
			PointsByDate: make(map[string]int64),
		},
	}

	updatedCustomers := customers

	suite.mockRuleRepo.EXPECT().
		GetActiveRules(ctx, gomock.Any()).
		DoAndReturn(func(ctx context.Context, branchCategories map[string][]string) ([]entity.Rule, error) {

			expectedMapping := map[string][]string{
				"BR0001": {"CT1001"},
				"BR0002": {"CT1002"},
			}
			suite.Equal(expectedMapping, branchCategories)
			return rules, nil
		})

	suite.mockCustRepo.EXPECT().
		GetCustomers(ctx, []string{"U000001", "U000002"}).
		Return(customers, nil)

	suite.mockCustRepo.EXPECT().
		UpdateBulkCustomers(ctx, gomock.Any()).
		DoAndReturn(func(ctx context.Context, updates []entity.UpdateCustomer) error {
			suite.Len(updates, 2)

			for _, update := range updates {
				if update.CustomerID == "U000001" {

					suite.Equal(int64(35), update.PointsToAdd)
				} else if update.CustomerID == "U000002" {

					suite.Equal(int64(27), update.PointsToAdd)
				}
			}
			return nil
		})

	suite.mockCustRepo.EXPECT().
		GetCustomers(ctx, []string{}).
		Return(updatedCustomers, nil)

	err := suite.service.ExecuteMultipleFiles(ctx, files)

	suite.NoError(err)
}

type CalculateBatchPointsTestSuite struct {
	suite.Suite
}

func TestCalculateBatchPointsTestSuite(t *testing.T) {
	suite.Run(t, new(CalculateBatchPointsTestSuite))
}

func (suite *CalculateBatchPointsTestSuite) TestCalculateBatchPoints_MultipleRulesPerRecord() {
	rules := []entity.Rule{
		{
			ID:       "RULE001",
			RuleType: entity.FixedPointRule,
			Conditions: entity.Conditions{
				MinAmount: decimal.NewFromFloat(50.0),
			},
			Reward: entity.Reward{Value: 10},
		},
		{
			ID:       "RULE002",
			RuleType: entity.PercentageRule,
			Conditions: entity.Conditions{
				MinAmount: decimal.NewFromFloat(50.0),
			},
			Reward: entity.Reward{Value: 5},
		},
	}

	records := PurchaseRecords{
		{
			CustomerID:      "U000001",
			ProductID:       "P001",
			BranchID:        "BR001",
			CategoryID:      "CT001",
			PurchasedAmount: decimal.NewFromFloat(100.0),
			PurchaseDate:    time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
		},
	}

	customers := entity.Customers{
		{
			CustomerID:   "U000001",
			Points:       0,
			Records:      []entity.Record{},
			PointsByDate: make(map[string]int64),
		},
	}

	result := calculateBatchPoints(rules, records, customers)
	suite.Len(result, 1)

	update := result[0]
	suite.Equal("U000001", update.CustomerID)
	suite.Equal(int64(15), update.PointsToAdd)
	suite.Equal(time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC), update.LastPurchaseDate)
	suite.Len(update.Records, 2)
}

func (suite *CalculateBatchPointsTestSuite) TestCalculateBatchPoints_ExistingCustomerPoints() {
	rules := []entity.Rule{
		{
			ID:       "RULE001",
			RuleType: entity.FixedPointRule,
			Conditions: entity.Conditions{
				MinAmount: decimal.NewFromFloat(50.0),
			},
			Reward: entity.Reward{Value: 10},
		},
	}

	records := PurchaseRecords{
		{
			CustomerID:      "U000001",
			ProductID:       "P001",
			BranchID:        "BR001",
			CategoryID:      "CT001",
			PurchasedAmount: decimal.NewFromFloat(100.0),
			PurchaseDate:    time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
		},
	}

	customers := entity.Customers{
		{
			CustomerID: "U000001",
			Points:     50,
			Records:    []entity.Record{},
			PointsByDate: map[string]int64{
				"2025-01-14": 50,
			},
		},
	}

	result := calculateBatchPoints(rules, records, customers)
	suite.Len(result, 1)

	update := result[0]
	suite.Equal("U000001", update.CustomerID)
	suite.Equal(int64(10), update.PointsToAdd)
	suite.Equal(int64(10), update.PointsByDate["2025-01-15"])
}

func (suite *CalculateBatchPointsTestSuite) TestCalculateBatchPoints_MultipleTransactionsSameDay() {
	rules := []entity.Rule{
		{
			ID:       "RULE001",
			RuleType: entity.FixedPointRule,
			Reward:   entity.Reward{Value: 10},
		},
	}
	records := PurchaseRecords{
		{
			CustomerID:      "U000001",
			PurchasedAmount: decimal.NewFromFloat(100.0),
			PurchaseDate:    time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			CustomerID:      "U000001",
			PurchasedAmount: decimal.NewFromFloat(200.0),
			PurchaseDate:    time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
		},
	}
	customers := entity.Customers{}

	result := calculateBatchPoints(rules, records, customers)
	suite.Len(result, 1)

	update := result[0]
	suite.Equal("U000001", update.CustomerID)
	suite.Equal(int64(20), update.PointsToAdd)
	suite.Equal(int64(20), update.PointsByDate["2025-01-15"])
}

type SaveFileTestSuite struct {
	suite.Suite
	tempDir string
}

func TestSaveFileTestSuite(t *testing.T) {
	suite.Run(t, new(SaveFileTestSuite))
}

func (suite *SaveFileTestSuite) SetupTest() {
	tempDir, err := os.MkdirTemp("", "test_save_")
	suite.NoError(err)
	suite.tempDir = tempDir
}

func (suite *SaveFileTestSuite) TearDownTest() {
	if suite.tempDir != "" {
		os.RemoveAll(suite.tempDir)
	}
}

func (suite *SaveFileTestSuite) TestSaveFile_Success() {
	customers := []entity.Customer{
		{
			CustomerID: "U000001",
			Points:     100,
			PointsByDate: map[string]int64{
				"2025-01-15": 100,
			},
		},
		{
			CustomerID: "U000002",
			Points:     200,
			PointsByDate: map[string]int64{
				"2025-01-15": 200,
			},
		},
	}

	filePath := filepath.Join(suite.tempDir, "output_%s.csv")
	dateString := "2025-01-15"

	err := saveFile(customers, filePath, dateString)

	suite.NoError(err)

	expectedFilePath := fmt.Sprintf(filePath, dateString)
	suite.FileExists(expectedFilePath)

	content, err := os.ReadFile(expectedFilePath)
	suite.NoError(err)

	contentStr := string(content)
	suite.Contains(contentStr, "customer_id,points,last_purchase_date")
	suite.Contains(contentStr, "U000001")
	suite.Contains(contentStr, "U000002")
}

func (suite *SaveFileTestSuite) TestSaveFile_SortedByPoints() {
	customers := []entity.Customer{
		{
			CustomerID: "U000001",
			Points:     100,
			PointsByDate: map[string]int64{
				"2025-01-15": 100,
			},
		},
		{
			CustomerID: "U000002",
			Points:     200,
			PointsByDate: map[string]int64{
				"2025-01-15": 200,
			},
		},
		{
			CustomerID: "U000003",
			Points:     50,
			PointsByDate: map[string]int64{
				"2025-01-15": 50,
			},
		},
	}

	filePath := filepath.Join(suite.tempDir, "output_%s.csv")
	dateString := "2025-01-15"

	err := saveFile(customers, filePath, dateString)

	suite.NoError(err)

	expectedFilePath := fmt.Sprintf(filePath, dateString)
	content, err := os.ReadFile(expectedFilePath)
	suite.NoError(err)

	lines := strings.Split(string(content), "\n")

	suite.Contains(lines[1], "U000002")
	suite.Contains(lines[2], "U000001")
	suite.Contains(lines[3], "U000003")
}

func (suite *SaveFileTestSuite) TestSaveFile_SortedByLastPurchaseDate() {
	customers := []entity.Customer{
		{
			CustomerID: "U000001",
			Points:     100,
			PointsByDate: map[string]int64{
				"2025-01-14": 100,
			},
		},
		{
			CustomerID: "U000002",
			Points:     100,
			PointsByDate: map[string]int64{
				"2025-01-15": 100,
			},
		},
	}

	filePath := filepath.Join(suite.tempDir, "output_%s.csv")
	dateString := "2025-01-15"

	err := saveFile(customers, filePath, dateString)

	suite.NoError(err)

	expectedFilePath := fmt.Sprintf(filePath, dateString)
	content, err := os.ReadFile(expectedFilePath)
	suite.NoError(err)

	lines := strings.Split(string(content), "\n")

	suite.Contains(lines[1], "U000002")
	suite.Contains(lines[2], "U000001")
}

func (suite *SaveFileTestSuite) TestSaveFile_InvalidFilePath() {
	customers := []entity.Customer{
		{
			CustomerID: "U000001",
			Points:     100,
			PointsByDate: map[string]int64{
				"2025-01-15": 100,
			},
		},
	}

	filePath := "/invalid/path/output_%s.csv"
	dateString := "2025-01-15"

	err := saveFile(customers, filePath, dateString)

	suite.Error(err)
	suite.IsType(&apperr.AppError{}, err)
}

func (suite *SaveFileTestSuite) TestSaveFile_EmptyCustomers() {
	customers := []entity.Customer{}

	filePath := filepath.Join(suite.tempDir, "output_%s.csv")
	dateString := "2025-01-15"

	err := saveFile(customers, filePath, dateString)

	suite.NoError(err)

	expectedFilePath := fmt.Sprintf(filePath, dateString)
	suite.FileExists(expectedFilePath)

	content, err := os.ReadFile(expectedFilePath)
	suite.NoError(err)

	contentStr := string(content)
	suite.Contains(contentStr, "customer_id,points,last_purchase_date")
}

type ExtendedValidateAndCalculatePointsTestSuite struct {
	suite.Suite
}

func TestExtendedValidateAndCalculatePointsTestSuite(t *testing.T) {
	suite.Run(t, new(ExtendedValidateAndCalculatePointsTestSuite))
}

func (suite *ExtendedValidateAndCalculatePointsTestSuite) TestValidateAndCalculatePoints_RatioRuleNilUnit() {
	rule := entity.Rule{
		RuleType: entity.RatioRule,
		Conditions: entity.Conditions{
			MinAmount: decimal.NewFromFloat(50.0),
		},
		Reward: entity.Reward{
			Value:     2,
			RatioUnit: nil,
		},
	}

	record := PurchaseRecord{
		CustomerID:      "U000001",
		ProductID:       "P001",
		BranchID:        "BR001",
		CategoryID:      "CT001",
		PurchasedAmount: decimal.NewFromFloat(100.0),
		PurchaseDate:    time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
	}

	customers := entity.Customers{}

	points, applied := validateAndCalculatePoints(rule, record, customers)

	suite.False(applied)
	suite.Equal(int64(0), points)
}

func (suite *ExtendedValidateAndCalculatePointsTestSuite) TestValidateAndCalculatePoints_RatioRuleZeroUnit() {
	ratioUnit := 0.0
	rule := entity.Rule{
		RuleType: entity.RatioRule,
		Conditions: entity.Conditions{
			MinAmount: decimal.NewFromFloat(50.0),
		},
		Reward: entity.Reward{
			Value:     2,
			RatioUnit: &ratioUnit,
		},
	}

	record := PurchaseRecord{
		CustomerID:      "U000001",
		ProductID:       "P001",
		BranchID:        "BR001",
		CategoryID:      "CT001",
		PurchasedAmount: decimal.NewFromFloat(100.0),
		PurchaseDate:    time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
	}

	customers := entity.Customers{}

	points, applied := validateAndCalculatePoints(rule, record, customers)

	suite.False(applied)
	suite.Equal(int64(0), points)
}

func (suite *ExtendedValidateAndCalculatePointsTestSuite) TestValidateAndCalculatePoints_RatioRuleNegativeUnit() {
	ratioUnit := -10.0
	rule := entity.Rule{
		RuleType: entity.RatioRule,
		Conditions: entity.Conditions{
			MinAmount: decimal.NewFromFloat(50.0),
		},
		Reward: entity.Reward{
			Value:     2,
			RatioUnit: &ratioUnit,
		},
	}

	record := PurchaseRecord{
		CustomerID:      "U000001",
		ProductID:       "P001",
		BranchID:        "BR001",
		CategoryID:      "CT001",
		PurchasedAmount: decimal.NewFromFloat(100.0),
		PurchaseDate:    time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
	}

	customers := entity.Customers{}

	points, applied := validateAndCalculatePoints(rule, record, customers)

	suite.False(applied)
	suite.Equal(int64(0), points)
}

func (suite *ExtendedValidateAndCalculatePointsTestSuite) TestValidateAndCalculatePoints_RatioRuleFloorCalculation() {
	ratioUnit := 30.0
	rule := entity.Rule{
		RuleType: entity.RatioRule,
		Conditions: entity.Conditions{
			MinAmount: decimal.NewFromFloat(50.0),
		},
		Reward: entity.Reward{
			Value:     5,
			RatioUnit: &ratioUnit,
		},
	}

	record := PurchaseRecord{
		CustomerID:      "U000001",
		ProductID:       "P001",
		BranchID:        "BR001",
		CategoryID:      "CT001",
		PurchasedAmount: decimal.NewFromFloat(175.0),
		PurchaseDate:    time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
	}

	customers := entity.Customers{}

	points, applied := validateAndCalculatePoints(rule, record, customers)

	suite.True(applied)
	suite.Equal(int64(25), points)
}

func (suite *ExtendedValidateAndCalculatePointsTestSuite) TestValidateAndCalculatePoints_UnknownRuleType() {
	rule := entity.Rule{
		RuleType: "UNKNOWN_RULE",
		Conditions: entity.Conditions{
			MinAmount: decimal.NewFromFloat(50.0),
		},
		Reward: entity.Reward{
			Value: 10,
		},
	}

	record := PurchaseRecord{
		CustomerID:      "U000001",
		ProductID:       "P001",
		BranchID:        "BR001",
		CategoryID:      "CT001",
		PurchasedAmount: decimal.NewFromFloat(100.0),
		PurchaseDate:    time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
	}

	customers := entity.Customers{}

	points, applied := validateAndCalculatePoints(rule, record, customers)

	suite.False(applied)
	suite.Equal(int64(0), points)
}

func (suite *ExtendedValidateAndCalculatePointsTestSuite) TestValidateAndCalculatePoints_CategoryIDMatching() {
	rule := entity.Rule{
		RuleType: entity.FixedPointRule,
		Conditions: entity.Conditions{
			MinAmount:  decimal.NewFromFloat(50.0),
			CategoryID: []string{"CT001", "CT002", "CT003"},
		},
		Reward: entity.Reward{
			Value: 10,
		},
	}

	record := PurchaseRecord{
		CustomerID:      "U000001",
		ProductID:       "P001",
		BranchID:        "BR001",
		CategoryID:      "CT002",
		PurchasedAmount: decimal.NewFromFloat(100.0),
		PurchaseDate:    time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
	}

	customers := entity.Customers{}

	points, applied := validateAndCalculatePoints(rule, record, customers)

	suite.True(applied)
	suite.Equal(int64(10), points)
}

func (suite *ExtendedValidateAndCalculatePointsTestSuite) TestValidateAndCalculatePoints_EmptyCategoryID() {
	rule := entity.Rule{
		RuleType: entity.FixedPointRule,
		Conditions: entity.Conditions{
			MinAmount:  decimal.NewFromFloat(50.0),
			CategoryID: []string{},
		},
		Reward: entity.Reward{
			Value: 10,
		},
	}

	record := PurchaseRecord{
		CustomerID:      "U000001",
		ProductID:       "P001",
		BranchID:        "BR001",
		CategoryID:      "ANY_CATEGORY",
		PurchasedAmount: decimal.NewFromFloat(100.0),
		PurchaseDate:    time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
	}

	customers := entity.Customers{}

	points, applied := validateAndCalculatePoints(rule, record, customers)

	suite.True(applied)
	suite.Equal(int64(10), points)
}

func (suite *ExtendedValidateAndCalculatePointsTestSuite) TestValidateAndCalculatePoints_EmptyBranchID() {
	rule := entity.Rule{
		RuleType: entity.FixedPointRule,
		Conditions: entity.Conditions{
			MinAmount: decimal.NewFromFloat(50.0),
			BranchID:  "",
		},
		Reward: entity.Reward{
			Value: 10,
		},
	}

	record := PurchaseRecord{
		CustomerID:      "U000001",
		ProductID:       "P001",
		BranchID:        "ANY_BRANCH",
		CategoryID:      "CT001",
		PurchasedAmount: decimal.NewFromFloat(100.0),
		PurchaseDate:    time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
	}

	customers := entity.Customers{}

	points, applied := validateAndCalculatePoints(rule, record, customers)

	suite.True(applied)
	suite.Equal(int64(10), points)
}

func (suite *ExtendedValidateAndCalculatePointsTestSuite) TestValidateAndCalculatePoints_CustomerNotFound() {
	rule := entity.Rule{
		RuleType: entity.FixedPointRule,
		Conditions: entity.Conditions{
			MinAmount: decimal.NewFromFloat(50.0),
		},
		Reward: entity.Reward{
			Value: 10,
		},
	}

	record := PurchaseRecord{
		CustomerID:      "UNKNOWN_CUSTOMER",
		ProductID:       "P001",
		BranchID:        "BR001",
		CategoryID:      "CT001",
		PurchasedAmount: decimal.NewFromFloat(100.0),
		PurchaseDate:    time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
	}

	customers := entity.Customers{
		{
			CustomerID: "U000001",
			Records:    []entity.Record{},
		},
	}

	points, applied := validateAndCalculatePoints(rule, record, customers)

	suite.True(applied)
	suite.Equal(int64(10), points)
}

func (suite *ExtendedValidateAndCalculatePointsTestSuite) TestValidateAndCalculatePoints_IsDuplicateRecord() {
	rule := entity.Rule{
		RuleType: entity.FixedPointRule,
		Conditions: entity.Conditions{
			MinAmount: decimal.NewFromFloat(50.0),
		},
		Reward: entity.Reward{
			Value: 10,
		},
	}

	record := PurchaseRecord{
		CustomerID:      "U000001",
		ProductID:       "P001",
		BranchID:        "BR001",
		CategoryID:      "CT001",
		PurchasedAmount: decimal.NewFromFloat(100.0),
		PurchaseDate:    time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
	}

	customers := entity.Customers{
		{
			CustomerID: "U000001",
			Records: []entity.Record{
				{
					ProductID:    "P001",
					BranchID:     "BR001",
					Amount:       decimal.NewFromFloat(100.0),
					PurchaseDate: time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
				},
			},
		},
	}

	points, applied := validateAndCalculatePoints(rule, record, customers)

	suite.False(applied)
	suite.Equal(int64(0), points)
}

func (suite *ExtendedValidateAndCalculatePointsTestSuite) TestValidateAndCalculatePoints_PercentageRuleZeroResult() {
	rule := entity.Rule{
		RuleType: entity.PercentageRule,
		Conditions: entity.Conditions{
			MinAmount: decimal.NewFromFloat(50.0),
		},
		Reward: entity.Reward{
			Value: 1,
		},
	}

	record := PurchaseRecord{
		CustomerID:      "U000001",
		ProductID:       "P001",
		BranchID:        "BR001",
		CategoryID:      "CT001",
		PurchasedAmount: decimal.NewFromFloat(99.99),
		PurchaseDate:    time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
	}

	customers := entity.Customers{}

	points, applied := validateAndCalculatePoints(rule, record, customers)

	suite.True(applied)
	suite.Equal(int64(0), points)
}

type ExtendedPurchaseRecordsTestSuite struct {
	suite.Suite
}

func TestExtendedPurchaseRecordsTestSuite(t *testing.T) {
	suite.Run(t, new(ExtendedPurchaseRecordsTestSuite))
}

func (suite *ExtendedPurchaseRecordsTestSuite) TestMapRecordsToBranchCategories_EmptyRecords() {
	records := PurchaseRecords{}

	result := records.mapRecordsToBranchCategories()

	suite.Empty(result)
}

func (suite *ExtendedPurchaseRecordsTestSuite) TestMapRecordsToBranchCategories_SingleBranchMultipleCategories() {
	records := PurchaseRecords{
		{
			BranchID:   "BR001",
			CategoryID: "CT001",
		},
		{
			BranchID:   "BR001",
			CategoryID: "CT002",
		},
		{
			BranchID:   "BR001",
			CategoryID: "CT003",
		},
	}

	result := records.mapRecordsToBranchCategories()

	suite.Len(result, 1)
	suite.Len(result["BR001"], 3)
	suite.Contains(result["BR001"], "CT001")
	suite.Contains(result["BR001"], "CT002")
	suite.Contains(result["BR001"], "CT003")
}

func (suite *ExtendedPurchaseRecordsTestSuite) TestMapRecordsToBranchCategories_DuplicateCategories() {
	records := PurchaseRecords{
		{
			BranchID:   "BR001",
			CategoryID: "CT001",
		},
		{
			BranchID:   "BR001",
			CategoryID: "CT001",
		},
		{
			BranchID:   "BR001",
			CategoryID: "CT002",
		},
	}

	result := records.mapRecordsToBranchCategories()

	suite.Len(result, 1)
	suite.Len(result["BR001"], 2)
	suite.Contains(result["BR001"], "CT001")
	suite.Contains(result["BR001"], "CT002")
}

func (suite *ExtendedPurchaseRecordsTestSuite) TestGetUniqueCustomerIDs_EmptyRecords() {
	records := PurchaseRecords{}

	result := records.getUniqueCustomerIDs()

	suite.Empty(result)
}

func (suite *ExtendedPurchaseRecordsTestSuite) TestGetUniqueCustomerIDs_DuplicateCustomers() {
	records := PurchaseRecords{
		{CustomerID: "U000001"},
		{CustomerID: "U000002"},
		{CustomerID: "U000001"},
		{CustomerID: "U000003"},
		{CustomerID: "U000002"},
	}

	result := records.getUniqueCustomerIDs()

	suite.Len(result, 3)
	suite.Contains(result, "U000001")
	suite.Contains(result, "U000002")
	suite.Contains(result, "U000003")
}

func (suite *ExtendedPurchaseRecordsTestSuite) TestGetUniqueRecords_ComplexDuplicates() {
	baseTime := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)

	records := PurchaseRecords{
		{
			CustomerID:      "U000001",
			ProductID:       "P001",
			BranchID:        "BR001",
			PurchasedAmount: decimal.NewFromFloat(100.0),
			PurchaseDate:    baseTime,
		},
		{
			CustomerID:      "U000001",
			ProductID:       "P001",
			BranchID:        "BR001",
			PurchasedAmount: decimal.NewFromFloat(100.0),
			PurchaseDate:    baseTime,
		},
		{
			CustomerID:      "U000001",
			ProductID:       "P001",
			BranchID:        "BR001",
			PurchasedAmount: decimal.NewFromFloat(100.01),
			PurchaseDate:    baseTime,
		},
		{
			CustomerID:      "U000001",
			ProductID:       "P001",
			BranchID:        "BR001",
			PurchasedAmount: decimal.NewFromFloat(100.0),
			PurchaseDate:    baseTime.Add(time.Hour),
		},
	}

	result := records.getUniqueRecords()

	suite.Len(result, 3)
}

func (suite *ExtendedPurchaseRecordsTestSuite) TestGetUniqueRecords_EmptyRecords() {
	records := PurchaseRecords{}

	result := records.getUniqueRecords()

	suite.Empty(result)
}

func (suite *ExtendedPurchaseRecordsTestSuite) TestGetUniqueRecords_AllUnique() {
	baseTime := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)

	records := PurchaseRecords{
		{
			CustomerID:      "U000001",
			ProductID:       "P001",
			BranchID:        "BR001",
			PurchasedAmount: decimal.NewFromFloat(100.0),
			PurchaseDate:    baseTime,
		},
		{
			CustomerID:      "U000002",
			ProductID:       "P001",
			BranchID:        "BR001",
			PurchasedAmount: decimal.NewFromFloat(100.0),
			PurchaseDate:    baseTime,
		},
		{
			CustomerID:      "U000001",
			ProductID:       "P002",
			BranchID:        "BR001",
			PurchasedAmount: decimal.NewFromFloat(100.0),
			PurchaseDate:    baseTime,
		},
	}

	result := records.getUniqueRecords()

	suite.Len(result, 3)
}
