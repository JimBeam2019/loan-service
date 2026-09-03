package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"loan-service/internal/domain"
	"loan-service/internal/repository"
	"loan-service/internal/utils"
	log "loan-service/pkg/logger"
	"loan-service/pkg/postgres"
	"math/big"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type LoanRepo struct {
	database              *postgres.Database
	tableName             string
	tableNameApproval     string
	tableNameInvestment   string
	tableNameDisbursement string
}

func NewLoanRepository(database *postgres.Database) repository.LoanRepository {
	return &LoanRepo{
		database:              database,
		tableName:             "loans",
		tableNameApproval:     "approvals",
		tableNameInvestment:   "investments",
		tableNameDisbursement: "disbursements",
	}
}

func (r *LoanRepo) Create(
	ctx context.Context,
	loan *domain.Loan,
	chanLoanId chan uuid.UUID,
	chanErr chan error,
) {
	defer close(chanLoanId)
	defer close(chanErr)

	if err := ctx.Err(); err != nil {
		chanLoanId <- uuid.Nil
		chanErr <- err
		return
	}
	if loan == nil {
		chanLoanId <- uuid.Nil
		chanErr <- errors.New("loan is required")
		return
	}

	r.database.Mtx.Lock()
	defer r.database.Mtx.Unlock()

	columns := []string{
		"borrower_id",
		"principal_amount",
		"rate",
		"return_of_investment",
	}

	principalAmount, rate := loan.PrincipalAmount(), loan.Rate()
	returnOfInvestment := loan.ReturnOfInvestment()

	// Send decimal values as strings so numeric(24,10) values are not
	// rounded through float64 on the way to PostgreSQL.
	values := []any{
		loan.BorrowerID(),
		principalAmount.Text('f', -1),
		rate.Text('f', -1),
		returnOfInvestment.Text('f', -1),
	}

	placeholders := make([]string, 0, len(values))

	for index := range values {
		placeholders = append(placeholders, fmt.Sprintf("$%d", index+1))
	}

	query := fmt.Sprintf(
		CreateQueryReturningId,
		r.tableName,
		strings.Join(columns, ", "),
		fmt.Sprintf("(%s)", strings.Join(placeholders, ", ")),
	)

	loanId, err := r.database.ExecQueryReturnUuid(ctx, query, values...)
	if errors.Is(err, pgx.ErrNoRows) {
		loanId = uuid.Nil
		err = domain.ErrDuplicateLoanFound
	}

	chanLoanId <- loanId
	chanErr <- err
}

func (r *LoanRepo) GetByID(
	ctx context.Context,
	id uuid.UUID,
	chanLoan chan domain.Loan,
	chanErr chan error,
) {
	defer close(chanLoan)
	defer close(chanErr)

	if err := ctx.Err(); err != nil {
		chanLoan <- domain.Loan{}
		chanErr <- err
		return
	}

	r.database.Mtx.RLock()
	defer r.database.Mtx.RUnlock()

	var (
		borrowerID                    uuid.UUID
		fPrincipalAmount, fRate, fRoi float64
		sqlAgreementLetterURL         sql.NullString
		state                         domain.LoanStateType
	)

	columns := []string{
		"borrower_id",
		"principal_amount",
		"rate",
		"return_of_investment",
		"agreement_letter_url",
		"state",
	}
	strColumns := strings.Join(columns, ", ")

	query := fmt.Sprintf(
		FetchQueryWhere,
		strColumns,
		r.tableName,
		"id = $1::uuid",
	)

	row := r.database.GetRowStmt(ctx, query, id)

	err := row.Scan(
		&borrowerID,
		&fPrincipalAmount,
		&fRate,
		&fRoi,
		&sqlAgreementLetterURL,
		&state,
	)
	if err != nil {
		chanLoan <- domain.Loan{}
		chanErr <- err
		return
	}

	principalAmount := big.NewFloat(fPrincipalAmount)
	rate := big.NewFloat(fRate)
	roi := big.NewFloat(fRoi)

	agreementLetterURL := ""
	if sqlAgreementLetterURL.Valid {
		agreementLetterURL = sqlAgreementLetterURL.String
	}

	loan := domain.SetLoanFromDB(
		id,
		borrowerID,
		*principalAmount,
		*rate,
		*roi,
		agreementLetterURL,
		state,
	)

	// Hydrate related records so GET and subsequent domain operations have
	// the complete aggregate rather than only the loans row.
	var approval domain.Approval
	err = r.database.GetRowStmt(ctx, `
		SELECT field_validator_employee_id, visit_proof_url, approved_at
		FROM approvals
		WHERE loan_id = $1
	`, id).Scan(
		&approval.FieldValidatorEmployeeID,
		&approval.VisitProofURL,
		&approval.ApprovalDate,
	)
	if err == nil {
		loan.RestoreApproval(approval)
	} else if !errors.Is(err, pgx.ErrNoRows) {
		chanLoan <- domain.Loan{}
		chanErr <- err
		return
	}

	rows, err := r.database.Connection.Query(ctx, `
		SELECT investor_id, amount, agreement_letter_url
		FROM investments
		WHERE loan_id = $1
		ORDER BY invested_at, id
	`, id)
	if err != nil {
		chanLoan <- domain.Loan{}
		chanErr <- err
		return
	}
	defer rows.Close()

	for rows.Next() {
		var (
			investorID          uuid.UUID
			fAmount             float64
			agreementLetterLink string
		)
		if err := rows.Scan(&investorID, &fAmount, &agreementLetterLink); err != nil {
			chanLoan <- domain.Loan{}
			chanErr <- err
			return
		}

		amount := big.NewFloat(fAmount)

		loan.RestoreInvestment(domain.Investment{
			InvestorID:         investorID,
			Amount:             *amount,
			AgreementLetterURL: agreementLetterLink,
		})
	}
	if err := rows.Err(); err != nil {
		chanLoan <- domain.Loan{}
		chanErr <- err
		return
	}

	var disbursement domain.Disbursement
	err = r.database.GetRowStmt(ctx, `
		SELECT signed_agreement_url, field_officer_employee_id, disbursed_at
		FROM disbursements
		WHERE loan_id = $1
	`, id).Scan(
		&disbursement.SignedAgreementURL,
		&disbursement.FieldOfficerEmployeeID,
		&disbursement.DisbursementDate,
	)
	if err == nil {
		loan.RestoreDisbursement(disbursement)
	} else if !errors.Is(err, pgx.ErrNoRows) {
		chanLoan <- domain.Loan{}
		chanErr <- err
		return
	}

	chanLoan <- *loan
	chanErr <- nil
}

func (r *LoanRepo) FindExistingLoan(
	ctx context.Context,
	loan *domain.Loan,
	chanLoan chan domain.Loan,
	chanErr chan error,
) {
	defer close(chanLoan)
	defer close(chanErr)

	if err := ctx.Err(); err != nil {
		chanLoan <- domain.Loan{}
		chanErr <- err
		return
	}

	r.database.Mtx.RLock()
	defer r.database.Mtx.RUnlock()

	var (
		id                    uuid.UUID
		sqlAgreementLetterURL sql.NullString
		state                 domain.LoanStateType
	)

	columns := []string{
		"id",
		"agreement_letter_url",
		"state",
	}
	strColumns := strings.Join(columns, ", ")

	whereConditions := []string{
		"borrower_id = $1::uuid",
		"principal_amount = $2::numeric",
		"rate = $3::numeric",
		"return_of_investment = $4::numeric",
		fmt.Sprintf("state <> '%s'::loan_state", domain.LoanStateDisbursed),
	}

	query := fmt.Sprintf(
		FetchQueryWhere,
		strColumns,
		r.tableName,
		strings.Join(whereConditions, string(SqlOperatorAnd)),
	)

	principalAmount := loan.PrincipalAmount()
	rate := loan.Rate()
	roi := loan.ReturnOfInvestment()

	row := r.database.GetRowStmt(
		ctx, query, loan.BorrowerID(), principalAmount.Text('f', -1), rate.Text('f', -1), roi.Text('f', -1),
	)

	err := row.Scan(
		&id,
		&sqlAgreementLetterURL,
		&state,
	)
	if err != nil {
		chanLoan <- domain.Loan{}
		chanErr <- err
		return
	}

	agreementLetterURL := ""
	if sqlAgreementLetterURL.Valid {
		agreementLetterURL = sqlAgreementLetterURL.String
	}

	curLoan := domain.SetLoanFromDB(
		id,
		loan.BorrowerID(),
		principalAmount,
		rate,
		roi,
		agreementLetterURL,
		state,
	)

	chanLoan <- *curLoan
	chanErr <- err
}

func (r *LoanRepo) Save(
	ctx context.Context,
	loan *domain.Loan,
	chanErr chan error,
) {
	defer close(chanErr)

	err := ctx.Err()
	if err != nil {
		chanErr <- err
		return
	}
	if loan == nil {
		chanErr <- errors.New("loan is required")
		return
	}

	r.database.Mtx.Lock()
	defer r.database.Mtx.Unlock()

	// Save is called after the domain transition. A fully-funded investment
	// changes the domain state to invested, so route by operation first.
	if loan.LastInvestment() != nil {
		chanErr <- r.investLoan(ctx, loan.ID, loan.LastInvestment())
		return
	}

	switch loan.Status() {
	case domain.LoanStateApproved:
		if approval := loan.Snapshot().Approval; approval != nil {
			chanErr <- r.approveLoan(ctx, loan.ID, approval)
			return
		}
	case domain.LoanStateDisbursed:
		if disbursement := loan.Snapshot().Disbursement; disbursement != nil {
			chanErr <- r.disburseLoan(ctx, loan.ID, disbursement)
			return
		}
	}

	chanErr <- domain.ErrInvalidStateTransition
}

func (r *LoanRepo) approveLoan(
	ctx context.Context,
	loanId uuid.UUID,
	approval *domain.Approval,
) error {
	logger := log.NewLogger()

	tx, err := r.database.Connection.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		logger.Error().Err(err).Msg("Failed to start transaction")
		return err
	}

	// Always rollback if the transaction hasn't been committed.
	// Rollback after Commit() is harmless.
	defer func() {
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil &&
			!errors.Is(rollbackErr, pgx.ErrTxClosed) {
			logger.Error().Err(rollbackErr).Msg("Failed to rollback transaction")
		}
	}()

	// 1. Lock the loan.
	var state domain.LoanStateType

	query := fmt.Sprintf(
		FetchQueryWhereForUpdate,
		"state",
		r.tableName,
		"id = $1::uuid",
	)

	err = tx.QueryRow(ctx, query, loanId).Scan(&state)
	if err != nil {
		logger.Error().Err(err).Str("query", query).Msg("Query error")
		return err
	}

	if state != domain.LoanStateProposed {
		return fmt.Errorf(
			"loan %s cannot be approved from state %s",
			loanId,
			state,
		)
	}

	// 2. Create the approval record.
	columnsApproval := []string{
		"loan_id",
		"field_validator_employee_id",
		"visit_proof_url",
		"approved_at",
	}

	valuesApproval := []any{
		loanId,
		approval.FieldValidatorEmployeeID,
		approval.VisitProofURL,
		approval.ApprovalDate,
	}

	placeholders := make([]string, 0, len(valuesApproval))

	for index := range valuesApproval {
		placeholders = append(placeholders, fmt.Sprintf("$%d", index+1))
	}

	query = fmt.Sprintf(
		CreateQuery,
		r.tableNameApproval,
		strings.Join(columnsApproval, ", "),
		fmt.Sprintf("(%s)", strings.Join(placeholders, ", ")),
	)
	_, err = tx.Exec(ctx, query, valuesApproval...)
	if err != nil {
		logger.Error().Err(err).Str("query", query).Msg("Query error")
		return err
	}

	// 3. Change the loan state.
	whereConditions := []string{
		"id = $1::uuid",
		fmt.Sprintf("state = '%s'::loan_state", domain.LoanStateProposed),
	}

	query = fmt.Sprintf(
		UpdateQueryWhere,
		r.tableName,
		fmt.Sprintf("state = '%s'::loan_state", domain.LoanStateApproved),
		strings.Join(whereConditions, string(SqlOperatorAnd)),
	)
	commandTag, err := tx.Exec(ctx, query, loanId)
	if err != nil {
		logger.Error().Err(err).Str("query", query).Msg("Query error")
		return err
	}

	if commandTag.RowsAffected() != 1 {
		return errors.New("loan state was not updated")
	}

	// 4. Commit everything atomically.
	return tx.Commit(ctx)
}

func (r *LoanRepo) investLoan(
	ctx context.Context,
	loanId uuid.UUID,
	investment *domain.Investment,
) error {
	logger := log.NewLogger()

	tx, err := r.database.Connection.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		logger.Error().Err(err).Msg("Failed to start transaction")
		return err
	}

	// Always rollback if the transaction hasn't been committed.
	// Rollback after Commit() is harmless.
	defer func() {
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil &&
			!errors.Is(rollbackErr, pgx.ErrTxClosed) {
			logger.Error().Err(rollbackErr).Msg("Failed to rollback transaction")
		}
	}()

	// 1. Lock the loan.
	var (
		state            domain.LoanStateType
		fPrincipalAmount float64
	)

	query := fmt.Sprintf(
		FetchQueryWhereForUpdate,
		"state, principal_amount",
		r.tableName,
		"id = $1::uuid",
	)

	err = tx.QueryRow(ctx, query, loanId).Scan(&state, &fPrincipalAmount)
	if err != nil {
		logger.Error().Err(err).Str("query", query).Msg("Query error")
		return err
	}

	if state != domain.LoanStateApproved {
		logger.Info().Str("state", string(state)).
			Msg("loan")
		return fmt.Errorf(
			"loan %s cannot be invested from state %s",
			loanId,
			state,
		)
	}

	// 2. Calculate existing investment.
	var sqlInvestedTotalAmount sql.NullFloat64

	query = fmt.Sprintf(
		FetchQueryWhere,
		"SUM(amount)",
		r.tableNameInvestment,
		"loan_id = $1::uuid",
	)

	err = tx.QueryRow(ctx, query, loanId).Scan(&sqlInvestedTotalAmount)
	if err != nil {
		logger.Error().Err(err).Str("query", query).Msg("Query error")
		return err
	}

	principalAmount := big.NewFloat(fPrincipalAmount)
	investedTotalAmount := big.NewFloat(0)
	if sqlInvestedTotalAmount.Valid {
		investedTotalAmount = new(big.Float).SetFloat64(sqlInvestedTotalAmount.Float64)
	}

	newInvestedTotalAmount := new(big.Float).Add(investedTotalAmount, &investment.Amount)

	logger.Info().Str("state", string(state)).
		Str("current total", investedTotalAmount.String()).
		Str("investment.Amount", investment.Amount.String()).
		Str("newInvestedTotalAmount", newInvestedTotalAmount.String()).
		Msg("loan")

	if utils.LessThan(principalAmount, newInvestedTotalAmount) {
		return domain.ErrInvestmentExceedsPrincipal
	}

	// 3. Create the investment record.
	columnsInvestment := []string{
		"loan_id",
		"investor_id",
		"amount",
		"agreement_letter_url",
	}

	valuesInvestment := []any{
		loanId,
		investment.InvestorID,
		investment.Amount.Text('f', -1),
		investment.AgreementLetterURL,
	}

	placeholders := make([]string, 0, len(valuesInvestment))

	for index := range valuesInvestment {
		placeholders = append(placeholders, fmt.Sprintf("$%d", index+1))
	}

	query = fmt.Sprintf(
		CreateQuery,
		r.tableNameInvestment,
		strings.Join(columnsInvestment, ", "),
		fmt.Sprintf("(%s)", strings.Join(placeholders, ", ")),
	)
	_, err = tx.Exec(ctx, query, valuesInvestment...)
	if err != nil {
		logger.Error().Err(err).Str("query", query).Msg("Query error")
		return err
	}

	// 4. Change the loan state when principalAmount == newInvestedTotalAmount.
	if utils.Equal(principalAmount, newInvestedTotalAmount) {
		whereConditions := []string{
			"id = $1::uuid",
			fmt.Sprintf("state = '%s'::loan_state", domain.LoanStateApproved),
		}

		query = fmt.Sprintf(
			UpdateQueryWhere,
			r.tableName,
			fmt.Sprintf("state = '%s'::loan_state", domain.LoanStateInvested),
			strings.Join(whereConditions, string(SqlOperatorAnd)),
		)
		commandTag, err := tx.Exec(ctx, query, loanId)
		if err != nil {
			logger.Error().Err(err).Str("query", query).Msg("Query error")
			return err
		}

		if commandTag.RowsAffected() != 1 {
			return errors.New("loan state was not updated")
		}
	}

	// 5. Commit everything atomically.
	return tx.Commit(ctx)
}

func (r *LoanRepo) disburseLoan(
	ctx context.Context,
	loanId uuid.UUID,
	disbursement *domain.Disbursement,
) error {

	logger := log.NewLogger()

	tx, err := r.database.Connection.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		logger.Error().Err(err).Msg("Failed to start transaction")
		return err
	}

	// Always rollback if the transaction hasn't been committed.
	// Rollback after Commit() is harmless.
	defer func() {
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil &&
			!errors.Is(rollbackErr, pgx.ErrTxClosed) {
			logger.Error().Err(rollbackErr).Msg("Failed to rollback transaction")
		}
	}()

	// 1. Lock the loan.
	var state domain.LoanStateType

	query := fmt.Sprintf(
		FetchQueryWhereForUpdate,
		"state",
		r.tableName,
		"id = $1::uuid",
	)

	err = tx.QueryRow(ctx, query, loanId).Scan(&state)
	if err != nil {
		logger.Error().Err(err).Str("query", query).Msg("Query error")
		return err
	}

	if state != domain.LoanStateInvested {
		return fmt.Errorf(
			"loan %s cannot be disbursed from state %s",
			loanId,
			state,
		)
	}

	// 2. Create the disbursement record.
	columnsDisbursement := []string{
		"loan_id",
		"signed_agreement_url",
		"field_officer_employee_id",
		"disbursed_at",
	}

	valuesDisbursement := []any{
		loanId,
		disbursement.SignedAgreementURL,
		disbursement.FieldOfficerEmployeeID,
		disbursement.DisbursementDate,
	}

	placeholders := make([]string, 0, len(valuesDisbursement))

	for index := range valuesDisbursement {
		placeholders = append(placeholders, fmt.Sprintf("$%d", index+1))
	}

	query = fmt.Sprintf(
		CreateQuery,
		r.tableNameDisbursement,
		strings.Join(columnsDisbursement, ", "),
		fmt.Sprintf("(%s)", strings.Join(placeholders, ", ")),
	)
	_, err = tx.Exec(ctx, query, valuesDisbursement...)
	if err != nil {
		logger.Error().Err(err).Str("query", query).Msg("Query error")
		return err
	}

	// 3. Change the loan state.
	whereConditions := []string{
		"id = $1::uuid",
		fmt.Sprintf("state = '%s'::loan_state", domain.LoanStateInvested),
	}

	query = fmt.Sprintf(
		UpdateQueryWhere,
		r.tableName,
		fmt.Sprintf("state = '%s'::loan_state", domain.LoanStateDisbursed),
		strings.Join(whereConditions, string(SqlOperatorAnd)),
	)
	commandTag, err := tx.Exec(ctx, query, loanId)
	if err != nil {
		logger.Error().Err(err).Str("query", query).Msg("Query error")
		return err
	}

	if commandTag.RowsAffected() != 1 {
		return errors.New("loan state was not updated")
	}

	// 4. Commit everything atomically.
	return tx.Commit(ctx)
}
