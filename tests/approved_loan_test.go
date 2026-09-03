package tests

import (
	"loan-service/internal/domain"
	"math/big"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoanCanBeApproved(t *testing.T) {
	testId, _ := uuid.NewUUID()

	loan, _ := domain.NewLoan(
		testId, *big.NewFloat(100_000), *big.NewFloat(0.05), *big.NewFloat(0.03),
	)

	approval := domain.Approval{
		FieldValidatorEmployeeID: testId,
		VisitProofURL:            "proof.com/file.pdf",
		ApprovalDate:             time.Now(),
	}

	err := loan.Approve(approval)

	require.NoError(t, err)
	assert.Equal(t, domain.LoanStateApproved, loan.Status())
}

func TestApprovalRequiresVisitProof(t *testing.T) {
	testId, _ := uuid.NewUUID()

	loan, _ := domain.NewLoan(
		testId, *big.NewFloat(100_000), *big.NewFloat(0.05), *big.NewFloat(0.03),
	)

	approval := domain.Approval{
		FieldValidatorEmployeeID: testId,
		VisitProofURL:            "",
		ApprovalDate:             time.Now(),
	}

	err := loan.Approve(approval)

	assert.ErrorIs(t, err, domain.ErrMissingVisitProof)
}

func TestApprovalRequiresValidatorAndDate(t *testing.T) {
	loan, err := domain.NewLoan(uuid.New(), *big.NewFloat(100_000), *big.NewFloat(0.05), *big.NewFloat(0.03))
	require.NoError(t, err)

	err = loan.Approve(domain.Approval{
		VisitProofURL: "https://example.com/proof.pdf",
		ApprovalDate:  time.Now(),
	})
	assert.ErrorIs(t, err, domain.ErrMissingValidator)

	err = loan.Approve(domain.Approval{
		FieldValidatorEmployeeID: uuid.New(),
		VisitProofURL:            "https://example.com/proof.pdf",
	})
	assert.ErrorIs(t, err, domain.ErrMissingApprovalDate)
}

func TestApprovedLoanCannotBeApprovedAgain(t *testing.T) {
	loan, err := domain.NewLoan(uuid.New(), *big.NewFloat(100), *big.NewFloat(0.05), *big.NewFloat(0.03))
	require.NoError(t, err)

	approval := domain.Approval{
		FieldValidatorEmployeeID: uuid.New(),
		VisitProofURL:            "https://example.com/proof.pdf",
		ApprovalDate:             time.Now(),
	}
	require.NoError(t, loan.Approve(approval))
	assert.Error(t, loan.Approve(approval))
}

func TestInvestmentRejectsZeroNegativeAndMissingInvestor(t *testing.T) {
	loan, err := domain.NewLoan(uuid.New(), *big.NewFloat(100), *big.NewFloat(0.05), *big.NewFloat(0.03))
	require.NoError(t, err)
	require.NoError(t, loan.Approve(domain.Approval{
		FieldValidatorEmployeeID: uuid.New(),
		VisitProofURL:            "https://example.com/proof.pdf",
		ApprovalDate:             time.Now(),
	}))

	assert.ErrorIs(t, loan.AddInvestment(domain.Investment{
		InvestorID: uuid.New(), Amount: *big.NewFloat(0),
	}), domain.ErrInvalidInvestmentAmount)

	assert.ErrorIs(t, loan.AddInvestment(domain.Investment{
		InvestorID: uuid.New(), Amount: *big.NewFloat(-1),
	}), domain.ErrInvalidInvestmentAmount)

	assert.ErrorIs(t, loan.AddInvestment(domain.Investment{
		Amount: *big.NewFloat(1),
	}), domain.ErrMissingInvestor)
}
