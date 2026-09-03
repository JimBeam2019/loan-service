package http

import (
	"errors"
	"loan-service/internal/domain"
	"math/big"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func parseTime(value string) (time.Time, error) {
	return time.Parse(time.RFC3339, value)
}

func parseFloat(value string) (big.Float, error) {
	var f big.Float
	_, _, err := f.Parse(value, 10)
	return f, err
}

func parseID(value string) (uuid.UUID, error) {
	return uuid.Parse(value)
}

func bad(ctx *gin.Context, err error) {
	status := http.StatusBadRequest
	if errors.Is(err, domain.ErrLoanNotFound) {
		status = http.StatusNotFound
	} else if errors.Is(err, domain.ErrDuplicateLoanFound) {
		status = http.StatusConflict
	}
	ctx.JSON(status, gin.H{"error": err.Error()})
}

func writeLoan(ctx *gin.Context, loan *domain.Loan) {
	ctx.JSON(http.StatusOK, loan.Snapshot())
}
