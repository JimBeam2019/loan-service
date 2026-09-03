package domain

import (
	"math/big"

	"github.com/google/uuid"
)

type Investment struct {
	InvestorID         uuid.UUID
	Amount             big.Float
	AgreementLetterURL string
}

func (i Investment) Validate() error {
	if i.InvestorID == uuid.Nil {
		return ErrMissingInvestor
	}
	if i.Amount.Sign() <= 0 {
		return ErrInvalidInvestmentAmount
	}
	return nil
}
