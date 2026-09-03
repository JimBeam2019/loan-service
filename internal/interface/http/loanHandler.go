package http

import (
	"loan-service/internal/domain"
	"loan-service/internal/usecase"
	"net/http"

	"github.com/gin-gonic/gin"
)

type LoanHandler struct{ uc *usecase.LoanUsecase }

func NewLoanHandler(r *gin.Engine, uc *usecase.LoanUsecase) {
	h := &LoanHandler{uc: uc}
	r.POST("/loans", h.Create)
	r.GET("/loans/:id", h.GetByID)
	r.POST("/loans/:id/approval", h.Approve)
	r.POST("/loans/:id/investments", h.AddInvestment)
	r.POST("/loans/:id/disbursement", h.Disburse)
}

type createLoanRequest struct {
	BorrowerID         string `json:"borrowerId" binding:"required"`
	PrincipalAmount    string `json:"principalAmount" binding:"required"`
	Rate               string `json:"rate" binding:"required"`
	ReturnOnInvestment string `json:"returnOnInvestment" binding:"required"`
}
type approvalRequest struct {
	FieldValidatorEmployeeID string `json:"fieldValidatorEmployeeId" binding:"required"`
	VisitProofURL            string `json:"visitProofUrl" binding:"required"`
	ApprovalDate             string `json:"approvalDate" binding:"required"`
}
type investmentRequest struct {
	InvestorID string `json:"investorId" binding:"required"`
	Amount     string `json:"amount" binding:"required"`
}
type disbursementRequest struct {
	SignedAgreementURL     string `json:"signedAgreementUrl" binding:"required"`
	FieldOfficerEmployeeID string `json:"fieldOfficerEmployeeId" binding:"required"`
	DisbursementDate       string `json:"disbursementDate" binding:"required"`
}

func (h *LoanHandler) Create(ctx *gin.Context) {
	var req createLoanRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		bad(ctx, err)
		return
	}
	borrowerID, err := parseID(req.BorrowerID)
	if err != nil {
		bad(ctx, err)
		return
	}
	principal, err := parseFloat(req.PrincipalAmount)
	if err != nil {
		bad(ctx, err)
		return
	}
	rate, err := parseFloat(req.Rate)
	if err != nil {
		bad(ctx, err)
		return
	}
	roi, err := parseFloat(req.ReturnOnInvestment)
	if err != nil {
		bad(ctx, err)
		return
	}
	loan, err := h.uc.Create(
		ctx.Request.Context(), borrowerID, principal, rate, roi,
	)
	if err != nil {
		bad(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, loan.Snapshot())
}

func (h *LoanHandler) GetByID(ctx *gin.Context) {
	id, err := parseID(ctx.Param("id"))
	if err != nil {
		bad(ctx, err)
		return
	}
	loan, err := h.uc.Get(ctx.Request.Context(), id)
	if err != nil {
		bad(ctx, err)
		return
	}
	writeLoan(ctx, loan)
}

func (h *LoanHandler) Approve(ctx *gin.Context) {
	id, err := parseID(ctx.Param("id"))
	if err != nil {
		bad(ctx, err)
		return
	}
	var req approvalRequest
	if err = ctx.ShouldBindJSON(&req); err != nil {
		bad(ctx, err)
		return
	}
	validatorID, err := parseID(req.FieldValidatorEmployeeID)
	if err != nil {
		bad(ctx, err)
		return
	}
	date, err := parseTime(req.ApprovalDate)
	if err != nil {
		bad(ctx, err)
		return
	}
	loan, err := h.uc.Approve(
		ctx.Request.Context(), id, domain.Approval{FieldValidatorEmployeeID: validatorID, VisitProofURL: req.VisitProofURL, ApprovalDate: date},
	)
	if err != nil {
		bad(ctx, err)
		return
	}
	writeLoan(ctx, loan)
}

func (h *LoanHandler) AddInvestment(ctx *gin.Context) {
	id, err := parseID(ctx.Param("id"))
	if err != nil {
		bad(ctx, err)
		return
	}
	var req investmentRequest
	if err = ctx.ShouldBindJSON(&req); err != nil {
		bad(ctx, err)
		return
	}
	investorID, err := parseID(req.InvestorID)
	if err != nil {
		bad(ctx, err)
		return
	}
	amount, err := parseFloat(req.Amount)
	if err != nil {
		bad(ctx, err)
		return
	}
	loan, err := h.uc.AddInvestment(
		ctx.Request.Context(), id, domain.Investment{InvestorID: investorID, Amount: amount},
	)
	if err != nil {
		bad(ctx, err)
		return
	}
	writeLoan(ctx, loan)
}

func (h *LoanHandler) Disburse(ctx *gin.Context) {
	id, err := parseID(ctx.Param("id"))
	if err != nil {
		bad(ctx, err)
		return
	}
	var req disbursementRequest
	if err = ctx.ShouldBindJSON(&req); err != nil {
		bad(ctx, err)
		return
	}
	officerID, err := parseID(req.FieldOfficerEmployeeID)
	if err != nil {
		bad(ctx, err)
		return
	}
	date, err := parseTime(req.DisbursementDate)
	if err != nil {
		bad(ctx, err)
		return
	}
	loan, err := h.uc.Disburse(
		ctx.Request.Context(), id, domain.Disbursement{SignedAgreementURL: req.SignedAgreementURL, FieldOfficerEmployeeID: officerID, DisbursementDate: date},
	)
	if err != nil {
		bad(ctx, err)
		return
	}
	writeLoan(ctx, loan)
}
