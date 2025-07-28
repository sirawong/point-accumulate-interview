package accumulatepoints

import (
	"testing"

	"github.com/sirawong/point-accumulate-interview/internal/domain/entity"
	"github.com/stretchr/testify/suite"
)

type UtilityFunctionsTestSuite struct {
	suite.Suite
}

func TestUtilityFunctionsTestSuite(t *testing.T) {
	suite.Run(t, new(UtilityFunctionsTestSuite))
}

func (suite *UtilityFunctionsTestSuite) TestEntityToCustomerPointRecord() {
	customers := []entity.Customer{
		{
			CustomerID: "U000001",
			Points:     250,
			PointsByDate: map[string]int64{
				"2025-01-15": 100,
				"2025-01-16": 150,
			},
		},
		{
			CustomerID: "U000002",
			Points:     200,
			PointsByDate: map[string]int64{
				"2025-01-14": 200,
			},
		},
	}

	targetDate := "2025-01-20"
	records := entityToCustomerPointRecord(customers, targetDate)

	suite.Len(records, 2)

	var cust001Record *CustomerPointRecord
	for _, record := range records {
		if record.CustomerID == "U000001" {
			cust001Record = record
			break
		}
	}

	suite.NotNil(cust001Record)
	suite.Equal("U000001", cust001Record.CustomerID)
	suite.Equal(int64(250), cust001Record.Points)
	suite.Equal("2025-01-16", cust001Record.LastPurchaseDate)
}

func (suite *UtilityFunctionsTestSuite) TestSumValuesOnOrBeforeDate() {
	pointsByDate := map[string]int64{
		"2025-01-10": 50,
		"2025-01-15": 100,
		"2025-01-20": 150,
	}

	value, lasttDate := sumValuesOnOrBeforeDate(pointsByDate, "2025-01-15")
	suite.Equal("2025-01-15", lasttDate)
	suite.Equal(int64(150), value)

	value, lasttDate = sumValuesOnOrBeforeDate(pointsByDate, "2025-01-18")
	suite.Equal("2025-01-15", lasttDate)
	suite.Equal(int64(150), value)

	value, lasttDate = sumValuesOnOrBeforeDate(pointsByDate, "2025-01-05")
	suite.Equal("", lasttDate)
	suite.Equal(int64(0), value)

	value, lasttDate = sumValuesOnOrBeforeDate(pointsByDate, "2025-01-25")
	suite.Equal("2025-01-20", lasttDate)
	suite.Equal(int64(300), value)
}
