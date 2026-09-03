package domain

// # Loan State Type
//   - proposed
//   - approved
//   - invested
//   - disbursed
type LoanStateType string

const (
	LoanStateProposed  LoanStateType = "proposed"
	LoanStateApproved  LoanStateType = "approved"
	LoanStateInvested  LoanStateType = "invested"
	LoanStateDisbursed LoanStateType = "disbursed"
)

type LoanState interface {
	Approve(loan *Loan, approval Approval) error
	AddInvestment(loan *Loan, investment Investment) error
	Disburse(loan *Loan, disbursement Disbursement) error
	Name() LoanStateType
}
