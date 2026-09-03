package domain

import "errors"

var (
	ErrInvalidStateTransition     = errors.New("invalid loan state transition")
	ErrLoanNotFound               = errors.New("loan not found")
	ErrDuplicateLoanFound         = errors.New("duplicate loan found")
	ErrMissingValidator           = errors.New("field validator is required")
	ErrMissingVisitProof          = errors.New("visit proof is required")
	ErrMissingApprovalDate        = errors.New("approval date is required")
	ErrMissingSignedAgreement     = errors.New("signed agreement is required")
	ErrMissingFieldOfficer        = errors.New("field officer is required")
	ErrMissingDisbursementDate    = errors.New("disbursement date is required")
	ErrInvestmentExceedsPrincipal = errors.New("investment exceeds loan principal")
	ErrInvalidInvestmentAmount    = errors.New("investment amount must be greater than zero")
	ErrMissingInvestor            = errors.New("investor ID is required")
	ErrInvalidRate                = errors.New("rate must be greater than or equal to zero")
	ErrInvalidReturnOnInvestment  = errors.New("return on investment must be greater than or equal to zero")
)
