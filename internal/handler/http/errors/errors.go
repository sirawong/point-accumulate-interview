package errors

import (
	"net/http"

	apperr "github.com/sirawong/point-accumulate-interview/internal/errors"

	"github.com/gin-gonic/gin"
)

func mapErrorToHTTPStatus(code string) int {
	switch code {
	case apperr.ErrNotFound.Code:
		return http.StatusNotFound
	case apperr.ErrInvalidArgument.Code:
		return http.StatusBadRequest
	case apperr.ErrUnauthenticated.Code:
		return http.StatusUnauthorized
	default:
		return http.StatusInternalServerError
	}
}

func RespondWithError(c *gin.Context, err error) {
	code := apperr.GetCode(err)
	httpStatus := mapErrorToHTTPStatus(code)
	c.JSON(httpStatus, gin.H{"error_code": code, "message": err.Error()})
}
