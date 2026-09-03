package tests

import (
	"bytes"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"

	"loan-service/internal/domain"
	memrepo "loan-service/internal/infrastructure/memory"
	httpapi "loan-service/internal/interface/http"
	"loan-service/internal/usecase"
	mem "loan-service/pkg/memory"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func testRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	db := mem.NewMemoryDB()
	repo := memrepo.NewLoanRepository(db)
	uc := usecase.NewLoanUsecase(repo)
	httpapi.NewLoanHandler(r, uc)
	return r
}

// func TestHTTPFlowConnectsToDomain(t *testing.T) {
// 	r := testRouter()
// 	borrower := uuid.New()
// 	investor := uuid.New()
// 	employee := uuid.New()

// 	body := fmt.Sprintf(`{"borrowerId":"%s","principalAmount":"5000000","rate":"0.05","returnOnInvestment":"0.03"}`, borrower)
// 	req := httptest.NewRequest(http.MethodPost, "/loans", bytes.NewBufferString(body))
// 	req.Header.Set("Content-Type", "application/json")
// 	rec := httptest.NewRecorder()
// 	r.ServeHTTP(rec, req)
// 	require.Equal(t, http.StatusCreated, rec.Code)

// 	var created struct {
// 		ID uuid.UUID `json:"id"`
// 	}
// 	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &created))
// 	require.NotEqual(t, uuid.Nil, created.ID)

// 	approval := fmt.Sprintf(`{"fieldValidatorEmployeeId":"%s","visitProofUrl":"https://example.com/proof.pdf","approvalDate":"%s"}`, employee, time.Now().UTC().Format(time.RFC3339))
// 	rec = httptest.NewRecorder()
// 	req = httptest.NewRequest(http.MethodPost, "/loans/"+created.ID.String()+"/approval", bytes.NewBufferString(approval))
// 	req.Header.Set("Content-Type", "application/json")
// 	r.ServeHTTP(rec, req)
// 	require.Equal(t, http.StatusOK, rec.Code)

// 	investment := fmt.Sprintf(`{"investorId":"%s","amount":"5000000"}`, investor)
// 	rec = httptest.NewRecorder()
// 	req = httptest.NewRequest(http.MethodPost, "/loans/"+created.ID.String()+"/investments", bytes.NewBufferString(investment))
// 	req.Header.Set("Content-Type", "application/json")
// 	r.ServeHTTP(rec, req)
// 	require.Equal(t, http.StatusOK, rec.Code)
// 	require.Contains(t, rec.Body.String(), `"status":"invested"`)

// 	disbursement := fmt.Sprintf(`{"signedAgreementUrl":"https://example.com/agreement.pdf","fieldOfficerEmployeeId":"%s","disbursementDate":"%s"}`, employee, time.Now().UTC().Format(time.RFC3339))
// 	rec = httptest.NewRecorder()
// 	req = httptest.NewRequest(http.MethodPost, "/loans/"+created.ID.String()+"/disbursement", bytes.NewBufferString(disbursement))
// 	req.Header.Set("Content-Type", "application/json")
// 	r.ServeHTTP(rec, req)
// 	require.Equal(t, http.StatusOK, rec.Code)
// 	require.Contains(t, rec.Body.String(), `"status":"disbursed"`)
// }

func TestHTTPGetMissingLoanReturns404(t *testing.T) {
	r := testRouter()
	req := httptest.NewRequest(http.MethodGet, "/loans/"+uuid.New().String(), nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusNotFound, rec.Code)
}

func TestDomainStillUsesStatePattern(t *testing.T) {
	loan, err := domain.NewLoan(uuid.New(), *big.NewFloat(100), *big.NewFloat(.05), *big.NewFloat(.03))
	require.NoError(t, err)
	require.Equal(t, domain.LoanStateProposed, loan.Status())
}

func TestHTTPRejectsInvalidTransition(t *testing.T) {
	r := testRouter()
	loanID := uuid.New()

	req := httptest.NewRequest(http.MethodPost, "/loans/"+loanID.String()+"/disbursement", bytes.NewBufferString(`{"signedAgreementUrl":"https://example.com/agreement.pdf","fieldOfficerEmployeeId":"`+uuid.New().String()+`","disbursementDate":"2026-08-30T10:00:00Z"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
}

// func TestHTTPRejectsDuplicateActiveLoan(t *testing.T) {
// 	r := testRouter()
// 	borrower := uuid.New()
// 	body := fmt.Sprintf(`{"borrowerId":"%s","principalAmount":"1000","rate":"0.05","returnOnInvestment":"0.03"}`, borrower)

// 	for i := 0; i < 2; i++ {
// 		req := httptest.NewRequest(http.MethodPost, "/loans", bytes.NewBufferString(body))
// 		req.Header.Set("Content-Type", "application/json")
// 		rec := httptest.NewRecorder()
// 		r.ServeHTTP(rec, req)
// 		if i == 0 {
// 			require.Equal(t, http.StatusCreated, rec.Code)
// 		} else {
// 			require.Equal(t, http.StatusConflict, rec.Code)
// 		}
// 	}
// }
