package domain

import (
	"errors"
	"loan-service/internal/utils"
	"math/big"
)

type ApprovedState struct{}

func (state *ApprovedState) Name() LoanStateType { return LoanStateApproved }

func (state *ApprovedState) Approve(loan *Loan, approval Approval) error {
	return errors.New("loan already approved")
}

func (state *ApprovedState) AddInvestment(loan *Loan, investment Investment) error {
	if err := investment.Validate(); err != nil {
		return err
	}

	total := loan.totalInvested()

	newTotal := new(big.Float).Add(&total, &investment.Amount)

	if utils.GreaterThan(newTotal, &loan.principalAmount) {
		return ErrInvestmentExceedsPrincipal
	}

	loan.investments = append(loan.investments, investment)
	loan.lastInvestment = &investment

	if utils.Equal(newTotal, &loan.principalAmount) {
		loan.transitionTo(&InvestedState{})
	}

	return nil
}

func (state *ApprovedState) Disburse(loan *Loan, disbursement Disbursement) error {
	return errors.New("loan not invested")
}
