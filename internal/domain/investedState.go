package domain

type InvestedState struct{}

func (state *InvestedState) Name() LoanStateType { return LoanStateInvested }

func (state *InvestedState) Approve(loan *Loan, approval Approval) error {
	return ErrInvalidStateTransition
}

func (state *InvestedState) AddInvestment(loan *Loan, investment Investment) error {
	return ErrInvalidStateTransition
}

func (state *InvestedState) Disburse(loan *Loan, disbursement Disbursement) error {
	if err := disbursement.Validate(); err != nil {
		return err
	}

	loan.disbursement = &disbursement
	loan.transitionTo(&DisbursedState{})

	return nil
}
