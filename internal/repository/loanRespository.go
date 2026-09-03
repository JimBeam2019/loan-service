package repository

import (
	"context"
	"loan-service/internal/domain"

	"github.com/google/uuid"
)

type LoanRepository interface {
	Create(
		ctx context.Context,
		loan *domain.Loan,
		chanLoanId chan uuid.UUID,
		chanErr chan error,
	)
	GetByID(
		ctx context.Context,
		id uuid.UUID,
		chanLoan chan domain.Loan,
		chanErr chan error,
	)
	FindExistingLoan(
		ctx context.Context,
		loan *domain.Loan,
		chanLoan chan domain.Loan,
		chanErr chan error,
	)
	Save(
		ctx context.Context,
		loan *domain.Loan,
		chanErr chan error,
	)
}
