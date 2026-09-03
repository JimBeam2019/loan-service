package domain

import (
	"github.com/google/uuid"
	"time"
)

type Disbursement struct {
	SignedAgreementURL     string    `json:"signedAgreementUrl"`
	FieldOfficerEmployeeID uuid.UUID `json:"fieldOfficerEmployeeId"`
	DisbursementDate       time.Time `json:"disbursementDate"`
}

func (d Disbursement) Validate() error {
	if d.SignedAgreementURL == "" {
		return ErrMissingSignedAgreement
	}
	if d.FieldOfficerEmployeeID == uuid.Nil {
		return ErrMissingFieldOfficer
	}
	if d.DisbursementDate.IsZero() {
		return ErrMissingDisbursementDate
	}
	return nil
}
