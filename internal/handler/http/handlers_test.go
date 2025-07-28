package http

import (
	"bytes"
	"context"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	apperr "github.com/sirawong/point-accumulate-interview/internal/errors"
	"github.com/sirawong/point-accumulate-interview/internal/services/accumulatepoints"
	"github.com/sirawong/point-accumulate-interview/internal/services/accumulatepoints/mocks"
	"github.com/sirawong/point-accumulate-interview/pkg/config"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

type AccumulatePointHandlerTestSuite struct {
	suite.Suite
	mockCtrl    *gomock.Controller
	mockService *mocks.MockAccumulatePointService
	cfg         *config.Config
	handler     *AccumulatePointHandler
	router      *gin.Engine
	tempDir     string
}

func TestAccumulatePointHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(AccumulatePointHandlerTestSuite))
}

func (suite *AccumulatePointHandlerTestSuite) SetupTest() {
	gin.SetMode(gin.TestMode)

	suite.mockCtrl = gomock.NewController(suite.T())
	suite.mockService = mocks.NewMockAccumulatePointService(suite.mockCtrl)

	tempDir, err := os.MkdirTemp("", "test_output_")
	suite.NoError(err)
	suite.tempDir = tempDir
	suite.cfg = &config.Config{
		FilePath: filepath.Join(tempDir, "output_%s.csv"),
	}
	suite.handler = NewAccumulatePointHandler(suite.mockService, suite.cfg)

	suite.router = gin.New()
	suite.router.POST("/upload", suite.handler.UploadCSV)
}

func (suite *AccumulatePointHandlerTestSuite) TearDownTest() {
	suite.mockCtrl.Finish()
}

func (suite *AccumulatePointHandlerTestSuite) TestUploadCSV_Success() {

	csvContent := "customer_id,product_id,category_id,category_name,branch_id,purchased_amount,currency\n" +
		"U000001,123191,CT1001,ELECTRONICS,BR3451,100.50,THB"

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part1, err := writer.CreatePart(map[string][]string{
		"Content-Disposition": {`form-data; name="csv_files"; filename="purchases_2025-01-15.csv"`},
		"Content-Type":        {"text/csv"},
	})
	suite.NoError(err)
	_, err = part1.Write([]byte(csvContent))
	suite.NoError(err)

	part2, err := writer.CreatePart(map[string][]string{
		"Content-Disposition": {`form-data; name="csv_files"; filename="purchases_2025-01-16.csv"`},
		"Content-Type":        {"text/csv"},
	})
	suite.NoError(err)
	_, err = part2.Write([]byte(csvContent))
	suite.NoError(err)

	err = writer.Close()
	suite.NoError(err)

	suite.mockService.EXPECT().
		ExecuteMultipleFiles(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, files []accumulatepoints.FileInput) error {
			suite.Len(files, 2)

			expectedDate1 := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
			expectedDate2 := time.Date(2025, 1, 16, 0, 0, 0, 0, time.UTC)
			suite.Equal(expectedDate1, files[0].PurchasedDate)
			suite.Equal(expectedDate2, files[1].PurchasedDate)
			return nil
		})

	req := httptest.NewRequest("POST", "/upload", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code)
	suite.Contains(w.Body.String(), "\"status\":\"ok\"")
}

func (suite *AccumulatePointHandlerTestSuite) TestUploadCSV_NoFiles() {

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	err := writer.Close()
	suite.NoError(err)

	req := httptest.NewRequest("POST", "/upload", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusBadRequest, w.Code)
	suite.Contains(w.Body.String(), "no files uploaded")
}

func (suite *AccumulatePointHandlerTestSuite) TestUploadCSV_InvalidFileExtension() {

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormFile("csv_files", "invalid_file.txt")
	suite.NoError(err)
	_, err = part.Write([]byte("some content"))
	suite.NoError(err)

	err = writer.Close()
	suite.NoError(err)

	req := httptest.NewRequest("POST", "/upload", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusBadRequest, w.Code)
	suite.Contains(w.Body.String(), "Invalid file type")
}

func (suite *AccumulatePointHandlerTestSuite) TestUploadCSV_ServiceError() {

	csvContent := "csv content"
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreatePart(map[string][]string{
		"Content-Disposition": {`form-data; name="csv_files"; filename="test_2025-01-15.csv"`},
		"Content-Type":        {"text/csv"},
	})
	suite.NoError(err)
	_, err = part.Write([]byte(csvContent))
	suite.NoError(err)

	err = writer.Close()
	suite.NoError(err)

	serviceError := apperr.ErrInternal.WithMessage("database connection failed")
	suite.mockService.EXPECT().
		ExecuteMultipleFiles(gomock.Any(), gomock.Any()).
		Return(serviceError)

	req := httptest.NewRequest("POST", "/upload", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusInternalServerError, w.Code)
	suite.Contains(w.Body.String(), "database connection failed")
}

type UtilityFunctionsTestSuite struct {
	suite.Suite
}

func TestUtilityFunctionsTestSuite(t *testing.T) {
	suite.Run(t, new(UtilityFunctionsTestSuite))
}

func (suite *UtilityFunctionsTestSuite) TestParseDateFromFilename_ValidFormats() {
	testCases := []struct {
		filename     string
		expectedDate time.Time
	}{
		{
			filename:     "purchases_2025-01-15.csv",
			expectedDate: time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			filename:     "2025-12-31_data.csv",
			expectedDate: time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC),
		},
		{
			filename:     "prefix_2025-06-15_suffix.csv",
			expectedDate: time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			filename:     "2025-01-01.csv",
			expectedDate: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.filename, func() {
			actualDate, err := parseDateFromFilename(tc.filename)
			suite.NoError(err)
			suite.Equal(tc.expectedDate, actualDate)
		})
	}
}

func (suite *UtilityFunctionsTestSuite) TestParseDateFromFilename_InvalidFormats() {
	testCases := []string{
		"invalid_filename.csv",
		"2025-1-15.csv",
		"25-01-15.csv",
		"2025/01/15.csv",
		"15-01-2025.csv",
		"no_date_here.csv",
		"2025-13-01.csv",
		"2025-01-32.csv",
	}

	for _, filename := range testCases {
		suite.Run(filename, func() {
			_, err := parseDateFromFilename(filename)
			suite.Error(err)
		})
	}
}

func (suite *UtilityFunctionsTestSuite) TestParseDateFromFilename_MultipleValidDates() {

	filename := "2025-01-15_backup_2025-12-31.csv"

	actualDate, err := parseDateFromFilename(filename)
	suite.NoError(err)

	expectedDate := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
	suite.Equal(expectedDate, actualDate)
}
