package tests

import (
	"context"
	"loan-service/internal/domain"
	memrepo "loan-service/internal/infrastructure/memory"
	mem "loan-service/pkg/memory"
	"math/big"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemoryRepositoryRejectsDuplicateActiveLoan(t *testing.T) {
	db := mem.NewMemoryDB()
	repo := memrepo.NewLoanRepository(db)
	ctx := context.Background()
	borrower := uuid.New()

	loan1, err := domain.NewLoan(borrower, *big.NewFloat(1000), *big.NewFloat(0.05), *big.NewFloat(0.03))
	require.NoError(t, err)
	ids := make(chan uuid.UUID, 1)
	errs := make(chan error, 1)
	repo.Create(ctx, loan1, ids, errs)
	require.NoError(t, <-errs)

	loan2, err := domain.NewLoan(borrower, *big.NewFloat(1000), *big.NewFloat(0.05), *big.NewFloat(0.03))
	require.NoError(t, err)
	ids = make(chan uuid.UUID, 1)
	errs = make(chan error, 1)
	repo.Create(ctx, loan2, ids, errs)

	assert.ErrorIs(t, <-errs, domain.ErrDuplicateLoanFound)
}
