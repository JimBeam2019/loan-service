package usecase

import (
	"context"
	"errors"
	"loan-service/internal/domain"
	"loan-service/internal/repository"
	log "loan-service/pkg/logger"
	"math/big"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type LoanUsecase struct{ repo repository.LoanRepository }

func NewLoanUsecase(r repository.LoanRepository) *LoanUsecase {
	return &LoanUsecase{repo: r}
}

func (u *LoanUsecase) Create(ctx context.Context, borrowerID uuid.UUID, principal, rate, roi big.Float) (*domain.Loan, error) {
	logger := log.NewLogger()

	loan, err := domain.NewLoan(borrowerID, principal, rate, roi)
	if err != nil {
		return nil, err
	}
	chanCurLoan := make(chan domain.Loan, 1)
	chanErrFind := make(chan error, 1)

	go u.repo.FindExistingLoan(ctx, loan, chanCurLoan, chanErrFind)

	curLoan, errFind := <-chanCurLoan, <-chanErrFind
	if errFind == nil {
		logger.Info().Interface("Loan", curLoan.Snapshot()).
			Msg("Found Loan")
		return nil, domain.ErrDuplicateLoanFound
	}

	if !errors.Is(errFind, pgx.ErrNoRows) {
		logger.Error().Err(errFind).Msg("DB Error")
		return nil, errFind
	}

	chanLoanId := make(chan uuid.UUID, 1)
	chanErr := make(chan error, 1)

	go u.repo.Create(ctx, loan, chanLoanId, chanErr)

	loanId, err := <-chanLoanId, <-chanErr
	if err != nil {
		return nil, err
	}
	loan.ID = loanId
	return loan, nil
}

func (u *LoanUsecase) Get(ctx context.Context, id uuid.UUID) (*domain.Loan, error) {
	chanLoan := make(chan domain.Loan, 1)
	chanErr := make(chan error, 1)

	go u.repo.GetByID(ctx, id, chanLoan, chanErr)

	loan, err := <-chanLoan, <-chanErr
	return &loan, err
}

func (u *LoanUsecase) Approve(ctx context.Context, id uuid.UUID, approval domain.Approval) (*domain.Loan, error) {
	loan, err := u.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if err = loan.Approve(approval); err != nil {
		return nil, err
	}

	chanErr := make(chan error, 1)

	go u.repo.Save(ctx, loan, chanErr)

	if err = <-chanErr; err != nil {
		return nil, err
	}
	return loan, nil
}

func (u *LoanUsecase) AddInvestment(ctx context.Context, id uuid.UUID, investment domain.Investment) (*domain.Loan, error) {
	loan, err := u.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if err = loan.AddInvestment(investment); err != nil {
		return nil, err
	}

	chanErr := make(chan error, 1)

	go u.repo.Save(ctx, loan, chanErr)

	if err = <-chanErr; err != nil {
		return nil, err
	}
	return loan, nil
}

func (u *LoanUsecase) Disburse(ctx context.Context, id uuid.UUID, disbursement domain.Disbursement) (*domain.Loan, error) {
	loan, err := u.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if err = loan.Disburse(disbursement); err != nil {
		return nil, err
	}

	chanErr := make(chan error, 1)

	go u.repo.Save(ctx, loan, chanErr)

	if err = <-chanErr; err != nil {
		return nil, err
	}
	return loan, nil
}
