package http

import (
	"archive/zip"
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"time"

	"github.com/gin-gonic/gin"
	apperr "github.com/sirawong/point-accumulate-interview/internal/errors"
	"github.com/sirawong/point-accumulate-interview/internal/handler/http/errors"
	"github.com/sirawong/point-accumulate-interview/internal/services/accumulatepoints"
	"github.com/sirawong/point-accumulate-interview/pkg/config"
	"github.com/sirawong/point-accumulate-interview/pkg/utils"
)

const (
	CSVFile         = ".csv"
	CSVtContentType = "text/csv"
	CSVKey          = "csv_files"
)

type AccumulatePointHandler struct {
	AccumulatePointSvc accumulatepoints.AccumulatePointService
	cfg                *config.Config
}

func NewAccumulatePointHandler(AccumulatePointSvc accumulatepoints.AccumulatePointService, cfg *config.Config) *AccumulatePointHandler {
	return &AccumulatePointHandler{AccumulatePointSvc: AccumulatePointSvc, cfg: cfg}
}

func (h AccumulatePointHandler) UploadCSV(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		errors.RespondWithError(c, apperr.ErrInvalidArgument.Wrap(err))
		return
	}

	files := form.File[CSVKey]
	if len(files) == 0 {
		errors.RespondWithError(c, apperr.ErrInvalidArgument.WithMessage("no files uploaded"))
		return
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].Filename < files[j].Filename
	})

	var fileReaders []accumulatepoints.FileInput
	for _, fileHeader := range files {

		if err = validateFileType(fileHeader); err != nil {
			errors.RespondWithError(c, err)
			return
		}

		file, err := fileHeader.Open()
		if err != nil {
			errors.RespondWithError(c, apperr.ErrInternal.Wrap(err))
			return
		}

		purchasedDate, err := parseDateFromFilename(fileHeader.Filename)
		if err != nil {
			errors.RespondWithError(c, err)
			return
		}

		fileReaders = append(fileReaders, accumulatepoints.FileInput{
			PurchasedDate: purchasedDate,
			Reader:        file,
		})
	}

	err = h.AccumulatePointSvc.ExecuteMultipleFiles(c.Request.Context(), fileReaders)
	if err != nil {
		errors.RespondWithError(c, err)
		return
	}

	zipWriter := zip.NewWriter(c.Writer)
	defer zipWriter.Close()
	if err = h.responseFile(fileReaders, zipWriter); err != nil {
		c.JSON(500, gin.H{"error": "Failed to create zip"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h AccumulatePointHandler) responseFile(fileReaders []accumulatepoints.FileInput, zipWriter *zip.Writer) error {
	for _, file := range fileReaders {
		date := file.PurchasedDate.Format(time.DateOnly)
		filePath := fmt.Sprintf(h.cfg.FilePath, date)

		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			continue
		}

		if err := utils.AddFileToZip(zipWriter, filePath, fmt.Sprintf("point-summary_%s.csv", date)); err != nil {
			return apperr.ErrInternal.Wrap(err)
		}
	}
	return nil
}

func validateFileType(fileHeader *multipart.FileHeader) error {
	extension := filepath.Ext(fileHeader.Filename)
	if extension != CSVFile {
		return apperr.ErrInvalidArgument.WithMessage("Invalid file type")
	}

	contentType := fileHeader.Header.Get("Content-Type")
	if contentType != CSVtContentType {
		return apperr.ErrInvalidArgument.WithMessage("Invalid content type")
	}

	return nil
}

func parseDateFromFilename(filename string) (time.Time, error) {
	var dateRegex = regexp.MustCompile(`\d{4}-\d{2}-\d{2}`)
	dateStr := dateRegex.FindString(filename)
	if dateStr == "" {
		errMsg := fmt.Errorf("date in YYYY-MM-DD format not found in filename: %s", filename)
		return time.Time{}, apperr.ErrInvalidArgument.Wrap(errMsg)
	}
	return time.Parse(time.DateOnly, dateStr)
}
