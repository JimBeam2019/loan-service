package domain

import (
	"errors"
	"loan-service/internal/utils"
	"math/big"

	"github.com/google/uuid"
)

type Loan struct {
	ID                 uuid.UUID `json:"id"`
	borrowerId         uuid.UUID
	principalAmount    big.Float
	rate               big.Float
	returnOfInvestment big.Float
	agreementLetterURL string
	investments        []Investment
	lastInvestment     *Investment
	approval           *Approval
	disbursement       *Disbursement
	state              LoanState
}

func NewLoan(borrowerID uuid.UUID, principalAmount, rate, roi big.Float) (*Loan, error) {
	if borrowerID == uuid.Nil {
		return nil, errors.New("borrower ID is required")
	}

	if !utils.GreaterThan(&principalAmount, big.NewFloat(0)) {
		return nil, errors.New("principal must be greater than zero")
	}

	if rate.Sign() < 0 {
		return nil, ErrInvalidRate
	}

	if roi.Sign() < 0 {
		return nil, ErrInvalidReturnOnInvestment
	}

	return &Loan{
		ID:                 uuid.New(),
		borrowerId:         borrowerID,
		principalAmount:    principalAmount,
		rate:               rate,
		returnOfInvestment: roi,
		investments:        make([]Investment, 0),
		state:              &ProposedState{},
	}, nil
}

func SetLoanFromDB(
	id, borrowerID uuid.UUID,
	principalAmount, rate, roi big.Float,
	agreementLetterURL string,
	state LoanStateType,
) *Loan {
	var loanState LoanState

	switch state {
	case LoanStateProposed:
		loanState = &ProposedState{}

	case LoanStateApproved:
		loanState = &ApprovedState{}

	case LoanStateInvested:
		loanState = &InvestedState{}

	case LoanStateDisbursed:
		loanState = &DisbursedState{}
	}

	return &Loan{
		ID:                 id,
		borrowerId:         borrowerID,
		principalAmount:    principalAmount,
		rate:               rate,
		returnOfInvestment: roi,
		agreementLetterURL: agreementLetterURL,
		investments:        make([]Investment, 0),
		state:              loanState,
	}
}

func (loan *Loan) transitionTo(state LoanState)    { loan.state = state }
func (loan *Loan) BorrowerID() uuid.UUID           { return loan.borrowerId }
func (loan *Loan) PrincipalAmount() big.Float      { return loan.principalAmount }
func (loan *Loan) Rate() big.Float                 { return loan.rate }
func (loan *Loan) ReturnOfInvestment() big.Float   { return loan.returnOfInvestment }
func (loan *Loan) LastInvestment() *Investment     { return loan.lastInvestment }
func (loan *Loan) Status() LoanStateType           { return loan.state.Name() }
func (loan *Loan) Approve(approval Approval) error { return loan.state.Approve(loan, approval) }
func (loan *Loan) AddInvestment(investment Investment) error {
	return loan.state.AddInvestment(loan, investment)
}
func (loan *Loan) Disburse(disbursement Disbursement) error {
	return loan.state.Disburse(loan, disbursement)
}
func (loan *Loan) TotalInvested() big.Float { return loan.totalInvested() }
func (loan *Loan) totalInvested() big.Float {
	total := big.NewFloat(0)
	for _, investment := range loan.investments {
		total = new(big.Float).Add(total, &investment.Amount)
	}
	return *total
}

// Snapshot exposes read-only business data for the HTTP layer without exposing mutable domain fields.
type Snapshot struct {
	ID                  uuid.UUID            `json:"id"`
	BorrowerID          uuid.UUID            `json:"borrowerId"`
	PrincipalAmount     string               `json:"principalAmount"`
	Rate                string               `json:"rate"`
	ReturnOnInvestment  string               `json:"returnOnInvestment"`
	AgreementLetterLink string               `json:"agreementLetterLink,omitempty"`
	Status              string               `json:"status"`
	TotalInvested       string               `json:"totalInvested"`
	Investments         []InvestmentSnapshot `json:"investments"`
	Approval            *Approval            `json:"approval,omitempty"`
	Disbursement        *Disbursement        `json:"disbursement,omitempty"`
}

type InvestmentSnapshot struct {
	InvestorID          uuid.UUID `json:"investorId"`
	Amount              string    `json:"amount"`
	AgreementLetterLink string    `json:"agreementLetterLink,omitempty"`
}

// RestoreApproval, RestoreInvestment and RestoreDisbursement hydrate an
// aggregate loaded from persistent storage without exposing mutable fields.
func (loan *Loan) RestoreApproval(approval Approval) {
	loan.approval = &approval
}

func (loan *Loan) RestoreInvestment(investment Investment) {
	loan.investments = append(loan.investments, investment)
}

func (loan *Loan) RestoreDisbursement(disbursement Disbursement) {
	loan.disbursement = &disbursement
}

func (loan *Loan) Snapshot() Snapshot {
	investments := make([]InvestmentSnapshot, 0, len(loan.investments))
	for _, i := range loan.investments {
		investments = append(investments, InvestmentSnapshot{
			InvestorID:          i.InvestorID,
			Amount:              i.Amount.Text('f', -1),
			AgreementLetterLink: i.AgreementLetterURL,
		})
	}
	loanTotalInvested := loan.TotalInvested()

	return Snapshot{
		ID:                  loan.ID,
		BorrowerID:          loan.borrowerId,
		PrincipalAmount:     loan.principalAmount.Text('f', -1),
		Rate:                loan.rate.Text('f', -1),
		ReturnOnInvestment:  loan.returnOfInvestment.Text('f', -1),
		AgreementLetterLink: loan.agreementLetterURL,
		Status:              string(loan.Status()),
		TotalInvested:       loanTotalInvested.Text('f', -1),
		Investments:         investments,
		Approval:            loan.approval,
		Disbursement:        loan.disbursement,
	}
}
