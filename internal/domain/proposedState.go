package domain

type ProposedState struct{}

func (state *ProposedState) Name() LoanStateType { return LoanStateProposed }

func (state *ProposedState) Approve(loan *Loan, approval Approval) error {
	if err := approval.Validate(); err != nil {
		return err
	}

	loan.approval = &approval
	loan.transitionTo(&ApprovedState{})

	return nil
}

func (state *ProposedState) AddInvestment(loan *Loan, investment Investment) error {
	return ErrInvalidStateTransition
}

func (state *ProposedState) Disburse(loan *Loan, disbursement Disbursement) error {
	return ErrInvalidStateTransition
}
