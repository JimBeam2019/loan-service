package domain

import (
	"github.com/google/uuid"
	"time"
)

type Approval struct {
	FieldValidatorEmployeeID uuid.UUID `json:"fieldValidatorEmployeeId"`
	VisitProofURL            string    `json:"visitProofUrl"`
	ApprovalDate             time.Time `json:"approvalDate"`
}

func (a Approval) Validate() error {
	if a.FieldValidatorEmployeeID == uuid.Nil {
		return ErrMissingValidator
	}
	if a.VisitProofURL == "" {
		return ErrMissingVisitProof
	}
	if a.ApprovalDate.IsZero() {
		return ErrMissingApprovalDate
	}
	return nil
}
