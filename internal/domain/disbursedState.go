package domain

type DisbursedState struct{}

func (state *DisbursedState) Name() LoanStateType { return LoanStateDisbursed }

func (state *DisbursedState) Approve(loan *Loan, approval Approval) error {
	return ErrInvalidStateTransition
}

func (state *DisbursedState) AddInvestment(loan *Loan, investment Investment) error {
	return ErrInvalidStateTransition
}

func (state *DisbursedState) Disburse(loan *Loan, disbursement Disbursement) error {
	return ErrInvalidStateTransition
}
