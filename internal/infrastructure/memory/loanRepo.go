package memory

import (
	"context"
	"errors"
	"loan-service/internal/domain"
	"loan-service/internal/repository"
	"loan-service/internal/utils"
	mem "loan-service/pkg/memory"

	"github.com/google/uuid"
)

type LoanRepo struct {
	memoryDB *mem.MemoryDB
}

func NewLoanRepository(memoryDB *mem.MemoryDB) repository.LoanRepository {
	return &LoanRepo{memoryDB: memoryDB}
}

func (r *LoanRepo) Create(
	ctx context.Context,
	loan *domain.Loan,
	chanLoanId chan uuid.UUID,
	chanErr chan error,
) {
	defer close(chanLoanId)
	defer close(chanErr)

	if err := ctx.Err(); err != nil {
		chanLoanId <- uuid.Nil
		chanErr <- err
		return
	}
	if loan == nil {
		chanLoanId <- uuid.Nil
		chanErr <- errors.New("loan is required")
		return
	}

	r.memoryDB.Mtx.Lock()
	defer r.memoryDB.Mtx.Unlock()

	if _, exists := r.memoryDB.Loans[loan.ID]; exists {
		chanLoanId <- uuid.Nil
		chanErr <- errors.New("loan already exists")
		return
	}

	// Keep the in-memory implementation consistent with the database's
	// active-loan uniqueness rule. The check and insert happen under the
	// same mutex, so concurrent requests cannot both create the loan.
	for _, curLoan := range r.memoryDB.Loans {
		if curLoan.BorrowerID() != loan.BorrowerID() ||
			curLoan.Status() == domain.LoanStateDisbursed {
			continue
		}
		principal, currentPrincipal := loan.PrincipalAmount(), curLoan.PrincipalAmount()
		rate, currentRate := loan.Rate(), curLoan.Rate()
		roi, currentROI := loan.ReturnOfInvestment(), curLoan.ReturnOfInvestment()
		if utils.Equal(&principal, &currentPrincipal) &&
			utils.Equal(&rate, &currentRate) &&
			utils.Equal(&roi, &currentROI) {
			chanLoanId <- uuid.Nil
			chanErr <- domain.ErrDuplicateLoanFound
			return
		}
	}

	r.memoryDB.Loans[loan.ID] = loan
	chanLoanId <- loan.ID
	chanErr <- nil
}

func (r *LoanRepo) GetByID(
	ctx context.Context,
	id uuid.UUID,
	chanLoan chan domain.Loan,
	chanErr chan error,
) {
	defer close(chanLoan)
	defer close(chanErr)

	if err := ctx.Err(); err != nil {
		chanLoan <- domain.Loan{}
		chanErr <- err
		return
	}

	r.memoryDB.Mtx.RLock()
	defer r.memoryDB.Mtx.RUnlock()

	loan, ok := r.memoryDB.Loans[id]
	if !ok {
		chanLoan <- domain.Loan{}
		chanErr <- domain.ErrLoanNotFound
		return
	}

	chanLoan <- *loan
	chanErr <- nil
}

func (r *LoanRepo) FindExistingLoan(
	ctx context.Context,
	loan *domain.Loan,
	chanLoan chan domain.Loan,
	chanErr chan error,
) {
	defer close(chanLoan)
	defer close(chanErr)

	if err := ctx.Err(); err != nil {
		chanLoan <- domain.Loan{}
		chanErr <- err
		return
	}

	r.memoryDB.Mtx.RLock()
	defer r.memoryDB.Mtx.RUnlock()

	for _, curLoan := range r.memoryDB.Loans {
		if loan.BorrowerID() != curLoan.BorrowerID() {
			continue
		}

		principalAmount := loan.PrincipalAmount()
		curPrincipalAmount := curLoan.PrincipalAmount()
		if !utils.Equal(&principalAmount, &curPrincipalAmount) {
			continue
		}

		rate, curRate := loan.Rate(), curLoan.Rate()
		if !utils.Equal(&rate, &curRate) {
			continue
		}

		roi, curROI := loan.ReturnOfInvestment(), curLoan.ReturnOfInvestment()
		if !utils.Equal(&roi, &curROI) {
			continue
		}

		if curLoan.Status() == domain.LoanStateDisbursed {
			continue
		}

		chanLoan <- *curLoan
		chanErr <- nil
		return
	}

	chanLoan <- domain.Loan{}
	chanErr <- domain.ErrLoanNotFound
}

func (r *LoanRepo) Save(
	ctx context.Context,
	loan *domain.Loan,
	chanErr chan error,
) {
	defer close(chanErr)

	if err := ctx.Err(); err != nil {
		chanErr <- err
		return
	}
	if loan == nil {
		chanErr <- errors.New("loan is required")
		return
	}

	r.memoryDB.Mtx.Lock()
	defer r.memoryDB.Mtx.Unlock()

	if _, exists := r.memoryDB.Loans[loan.ID]; !exists {
		chanErr <- domain.ErrLoanNotFound
		return
	}
	r.memoryDB.Loans[loan.ID] = loan

	chanErr <- nil
}
