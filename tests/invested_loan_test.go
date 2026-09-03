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

func TestLoanBecomesInvestedWhenFullyFunded(t *testing.T) {
	testId, _ := uuid.NewUUID()

	loan, _ := domain.NewLoan(
		testId, *big.NewFloat(5_000_000), *big.NewFloat(0.05), *big.NewFloat(0.03),
	)

	approval := domain.Approval{
		FieldValidatorEmployeeID: testId,
		VisitProofURL:            "proof.com/file.pdf",
		ApprovalDate:             time.Now(),
	}

	require.NoError(t, loan.Approve(approval))

	investment1 := domain.Investment{
		InvestorID: testId,
		Amount:     *big.NewFloat(2_000_000),
	}

	require.NoError(t, loan.AddInvestment(investment1))

	assert.Equal(t, domain.LoanStateApproved, loan.Status())

	investment2 := domain.Investment{
		InvestorID: testId,
		Amount:     *big.NewFloat(3_000_000),
	}

	require.NoError(t, loan.AddInvestment(investment2))

	assert.Equal(t, domain.LoanStateInvested, loan.Status())
}

func TestInvestmentCannotExceedPrincipal(t *testing.T) {
	testId, _ := uuid.NewUUID()

	loan, _ := domain.NewLoan(
		testId, *big.NewFloat(5_000_000), *big.NewFloat(0.05), *big.NewFloat(0.03),
	)

	approval := domain.Approval{
		FieldValidatorEmployeeID: testId,
		VisitProofURL:            "proof.com/file.pdf",
		ApprovalDate:             time.Now(),
	}

	require.NoError(t, loan.Approve(approval))

	investment := domain.Investment{
		InvestorID: testId,
		Amount:     *big.NewFloat(5_000_001),
	}

	assert.ErrorIs(
		t,
		loan.AddInvestment(investment),
		domain.ErrInvestmentExceedsPrincipal,
	)

	totalInvested := loan.TotalInvested()

	assert.Equal(t, "0", totalInvested.String())
}

func TestInvestmentExactlyAtPrincipalMovesToInvested(t *testing.T) {
	loan, err := domain.NewLoan(uuid.New(), *big.NewFloat(100), *big.NewFloat(0.05), *big.NewFloat(0.03))
	require.NoError(t, err)
	require.NoError(t, loan.Approve(domain.Approval{
		FieldValidatorEmployeeID: uuid.New(),
		VisitProofURL:            "https://example.com/proof.pdf",
		ApprovalDate:             time.Now(),
	}))

	require.NoError(t, loan.AddInvestment(domain.Investment{
		InvestorID: uuid.New(), Amount: *big.NewFloat(100),
	}))

	assert.Equal(t, domain.LoanStateInvested, loan.Status())
	totalInvested := loan.TotalInvested()
	assert.Equal(t, "100", totalInvested.String())
}

func TestInvestmentCannotBeAddedAfterFullyFunded(t *testing.T) {
	loan, err := domain.NewLoan(uuid.New(), *big.NewFloat(100), *big.NewFloat(0.05), *big.NewFloat(0.03))
	require.NoError(t, err)
	require.NoError(t, loan.Approve(domain.Approval{
		FieldValidatorEmployeeID: uuid.New(),
		VisitProofURL:            "https://example.com/proof.pdf",
		ApprovalDate:             time.Now(),
	}))
	require.NoError(t, loan.AddInvestment(domain.Investment{
		InvestorID: uuid.New(), Amount: *big.NewFloat(100),
	}))

	err = loan.AddInvestment(domain.Investment{
		InvestorID: uuid.New(), Amount: *big.NewFloat(1),
	})
	assert.ErrorIs(t, err, domain.ErrInvalidStateTransition)
}

func TestApprovedLoanCanReceiveMultipleInvestors(t *testing.T) {
	loan, err := domain.NewLoan(uuid.New(), *big.NewFloat(100), *big.NewFloat(0.05), *big.NewFloat(0.03))
	require.NoError(t, err)
	require.NoError(t, loan.Approve(domain.Approval{
		FieldValidatorEmployeeID: uuid.New(),
		VisitProofURL:            "https://example.com/proof.pdf",
		ApprovalDate:             time.Now(),
	}))

	require.NoError(t, loan.AddInvestment(domain.Investment{InvestorID: uuid.New(), Amount: *big.NewFloat(40)}))
	require.NoError(t, loan.AddInvestment(domain.Investment{InvestorID: uuid.New(), Amount: *big.NewFloat(60)}))

	totalInvested := loan.TotalInvested()

	assert.Equal(t, "100", totalInvested.String())
	assert.Len(t, loan.Snapshot().Investments, 2)
	assert.Equal(t, domain.LoanStateInvested, loan.Status())
}

func TestInvestedLoanCanBeDisbursed(t *testing.T) {
	loan, err := domain.NewLoan(uuid.New(), *big.NewFloat(100), *big.NewFloat(0.05), *big.NewFloat(0.03))
	require.NoError(t, err)
	require.NoError(t, loan.Approve(domain.Approval{
		FieldValidatorEmployeeID: uuid.New(),
		VisitProofURL:            "https://example.com/proof.pdf",
		ApprovalDate:             time.Now(),
	}))
	require.NoError(t, loan.AddInvestment(domain.Investment{
		InvestorID: uuid.New(), Amount: *big.NewFloat(100),
	}))

	disbursement := domain.Disbursement{
		SignedAgreementURL:     "https://example.com/signed-agreement.pdf",
		FieldOfficerEmployeeID: uuid.New(),
		DisbursementDate:       time.Now(),
	}
	require.NoError(t, loan.Disburse(disbursement))
	assert.Equal(t, domain.LoanStateDisbursed, loan.Status())
}

func TestDisbursementRequiresAllFields(t *testing.T) {
	loan, err := domain.NewLoan(uuid.New(), *big.NewFloat(100), *big.NewFloat(0.05), *big.NewFloat(0.03))
	require.NoError(t, err)
	require.NoError(t, loan.Approve(domain.Approval{
		FieldValidatorEmployeeID: uuid.New(),
		VisitProofURL:            "https://example.com/proof.pdf",
		ApprovalDate:             time.Now(),
	}))
	require.NoError(t, loan.AddInvestment(domain.Investment{
		InvestorID: uuid.New(), Amount: *big.NewFloat(100),
	}))

	err = loan.Disburse(domain.Disbursement{
		FieldOfficerEmployeeID: uuid.New(),
		DisbursementDate:       time.Now(),
	})
	assert.ErrorIs(t, err, domain.ErrMissingSignedAgreement)
}

func TestInvestedLoanCannotBeApprovedOrInvested(t *testing.T) {
	loan, err := domain.NewLoan(uuid.New(), *big.NewFloat(100), *big.NewFloat(0.05), *big.NewFloat(0.03))
	require.NoError(t, err)
	require.NoError(t, loan.Approve(domain.Approval{
		FieldValidatorEmployeeID: uuid.New(),
		VisitProofURL:            "https://example.com/proof.pdf",
		ApprovalDate:             time.Now(),
	}))
	require.NoError(t, loan.AddInvestment(domain.Investment{
		InvestorID: uuid.New(), Amount: *big.NewFloat(100),
	}))

	assert.ErrorIs(t, loan.AddInvestment(domain.Investment{InvestorID: uuid.New(), Amount: *big.NewFloat(1)}), domain.ErrInvalidStateTransition)
	assert.ErrorIs(t, loan.Approve(domain.Approval{FieldValidatorEmployeeID: uuid.New(), VisitProofURL: "x", ApprovalDate: time.Now()}), domain.ErrInvalidStateTransition)
}
