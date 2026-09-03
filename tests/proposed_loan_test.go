package tests

import (
	"loan-service/internal/domain"
	"math/big"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoanStartsProposed(t *testing.T) {
	loan, err := domain.NewLoan(uuid.New(), *big.NewFloat(100_000), *big.NewFloat(0.05), *big.NewFloat(0.03))
	require.NoError(t, err)
	assert.Equal(t, domain.LoanStateProposed, loan.Status())
}

func TestNewLoanRejectsInvalidValues(t *testing.T) {
	tests := []struct {
		name                 string
		borrower             uuid.UUID
		principal, rate, roi *big.Float
		expected             error
	}{
		{"missing borrower", uuid.Nil, big.NewFloat(100), big.NewFloat(0.05), big.NewFloat(0.03), nil},
		{"zero principal", uuid.New(), big.NewFloat(0), big.NewFloat(0.05), big.NewFloat(0.03), nil},
		{"negative principal", uuid.New(), big.NewFloat(-1), big.NewFloat(0.05), big.NewFloat(0.03), nil},
		{"negative rate", uuid.New(), big.NewFloat(100), big.NewFloat(-0.01), big.NewFloat(0.03), domain.ErrInvalidRate},
		{"negative ROI", uuid.New(), big.NewFloat(100), big.NewFloat(0.05), big.NewFloat(-0.01), domain.ErrInvalidReturnOnInvestment},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := domain.NewLoan(tt.borrower, *tt.principal, *tt.rate, *tt.roi)
			require.Error(t, err)
			if tt.expected != nil {
				assert.ErrorIs(t, err, tt.expected)
			}
		})
	}
}

func TestProposedLoanCannotBeInvestedOrDisbursed(t *testing.T) {
	loan, err := domain.NewLoan(uuid.New(), *big.NewFloat(100), *big.NewFloat(0.05), *big.NewFloat(0.03))
	require.NoError(t, err)

	assert.ErrorIs(t, loan.AddInvestment(domain.Investment{
		InvestorID: uuid.New(), Amount: *big.NewFloat(10),
	}), domain.ErrInvalidStateTransition)

	assert.ErrorIs(t, loan.Disburse(domain.Disbursement{}), domain.ErrInvalidStateTransition)
}
