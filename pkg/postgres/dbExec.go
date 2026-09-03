package postgres

import (
	"context"
	"errors"
	"fmt"
	log "loan-service/pkg/logger"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// # Connect to database
func ConnectPostgres(
	ctx context.Context,
	host, port, database, user, password string,
) (*pgxpool.Pool, error) {
	logger := log.NewLogger()

	connString := fmt.Sprintf(
		"postgresql://%s:%s@%s:%s/%s?sslmode=disable",
		user,
		password,
		host,
		port,
		database,
	)

	logger.Info().Str("host", host).Str("port", port).
		Str("database", database).Str("user", user).
		Msg("Connecting to PostgreSQL")

	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		logger.Error().Err(err).Msg("Unable to open database connection")
		return nil, err
	}

	config.MaxConns = 25
	config.MinConns = 5
	config.MaxConnIdleTime = 25
	config.MaxConnLifetime = 5 * time.Minute

	db, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		logger.Fatal().Err(err).Str("connString", connString).
			Msg("Postgres open connection Failed")
	}

	if err = db.Ping(ctx); err != nil {
		db.Close()
		logger.Error().Err(err).Msg("Postgres ping failed")
		return nil, err
	}

	return db, nil
}

func ClosePostgres(db *pgxpool.Pool) {
	if db != nil {
		db.Close()
	}
}

// -----------------------------------------------------------------------------
// Query helpers
// -----------------------------------------------------------------------------

// # GetRowsStmt executes a query and returns rows.
//
// The function name is kept for compatibility with the existing code,
// although pgx does not require a prepared statement here.
func (db *Database) GetRowsStmt(ctx context.Context, query string, args ...any) (
	pgx.Rows, error,
) {
	logger := log.NewLogger()

	rows, err := db.Connection.Query(ctx, query, args...)
	if err != nil {
		logger.Error().Err(err).Str("query", query).
			Msg("Failed to execute query")
		return nil, err
	}

	return rows, nil
}

// # GetRowStmt executes a query expected to return a single row.
func (db *Database) GetRowStmt(
	ctx context.Context,
	query string,
	args ...any,
) pgx.Row {
	return db.Connection.QueryRow(ctx, query, args...)
}

// -----------------------------------------------------------------------------
// Raw query
// -----------------------------------------------------------------------------

// # ExecRawQueryReturnError executes a raw SQL query.
func (db *Database) ExecRawQueryReturnError(ctx context.Context, query string) error {
	logger := log.NewLogger()

	_, err := db.Connection.Exec(ctx, query)
	if err != nil {
		logger.Error().Err(err).Str("query", query).
			Msg("Failed to execute query")

		return err
	}

	return nil
}

// -----------------------------------------------------------------------------
// Transaction helpers
// -----------------------------------------------------------------------------

// # ExecTransactionReturnError executes a statement inside a transaction.
func (db *Database) ExecTransactionReturnError(
	ctx context.Context,
	query string,
	args ...any,
) error {

	logger := log.NewLogger()

	tx, err := db.Connection.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		logger.Error().Err(err).Str("query", query).
			Msg("Failed to start transaction")

		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		}
	}()

	commandTag, err := tx.Exec(ctx, query, args...)
	if err != nil {
		logger.Error().Err(err).Str("query", query).
			Msg("Failed to execute statement")

		return err
	}

	if commandTag.RowsAffected() == 0 {
		return errors.New("no rows affected")
	}

	if err = tx.Commit(ctx); err != nil {
		logger.Error().Err(err).Str("query", query).
			Msg("Failed to commit transaction")

		return err
	}

	return nil
}

// ExecTransactionReturnRes executes a statement inside a transaction
// and returns the PostgreSQL command tag.
func (db *Database) ExecTransactionReturnRes(
	ctx context.Context,
	query string,
	args ...any,
) (pgconn.CommandTag, error) {

	logger := log.NewLogger()

	tx, err := db.Connection.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		logger.Error().Err(err).Str("query", query).
			Msg("Failed to start transaction")

		return pgconn.CommandTag{}, err
	}

	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		}
	}()

	res, err := tx.Exec(ctx, query, args...)
	if err != nil {
		logger.Error().Err(err).Str("query", query).
			Msg("Failed to execute statement")

		return pgconn.CommandTag{}, err
	}

	if err = tx.Commit(ctx); err != nil {
		logger.Error().Err(err).Str("query", query).
			Msg("Failed to commit transaction")

		return pgconn.CommandTag{}, err
	}

	return res, nil
}

// -----------------------------------------------------------------------------
// Return integer IDs
// -----------------------------------------------------------------------------

// ExecTransactionReturnIds executes a query returning integer IDs.
func (db *Database) ExecTransactionReturnIds(
	ctx context.Context,
	query string,
	args ...any,
) ([]int, error) {

	logger := log.NewLogger()

	tx, err := db.Connection.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		logger.Error().Err(err).Str("query", query).
			Msg("Failed to start transaction")

		return nil, err
	}

	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		}
	}()

	rows, err := tx.Query(ctx, query, args...)
	if err != nil {
		logger.Error().Err(err).Str("query", query).
			Msg("Failed to execute query")

		return nil, err
	}

	defer rows.Close()

	var ids []int

	for rows.Next() {
		var id int

		if err = rows.Scan(&id); err != nil {
			logger.Error().Err(err).Str("query", query).
				Msg("Failed to scan returned ID")

			return nil, err
		}

		ids = append(ids, id)
	}

	if err = rows.Err(); err != nil {
		logger.Error().Err(err).Str("query", query).
			Msg("Error while reading returned IDs")

		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		logger.Error().Err(err).Str("query", query).
			Msg("Failed to commit transaction")

		return nil, err
	}

	return ids, nil
}

// -----------------------------------------------------------------------------
// Return UUIDs
// -----------------------------------------------------------------------------

// ExecTransactionReturnUuids executes a query returning UUIDs.
func (db *Database) ExecTransactionReturnUuids(
	ctx context.Context,
	query string,
	args ...any,
) ([]uuid.UUID, error) {

	logger := log.NewLogger()

	tx, err := db.Connection.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		logger.Error().Err(err).Str("query", query).
			Msg("Failed to start transaction")

		return nil, err
	}

	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		}
	}()

	rows, err := tx.Query(ctx, query, args...)
	if err != nil {
		logger.Error().Err(err).Str("query", query).
			Msg("Failed to execute query")

		return nil, err
	}

	defer rows.Close()

	var ids []uuid.UUID

	for rows.Next() {
		var id uuid.UUID

		if err = rows.Scan(&id); err != nil {
			logger.Error().Err(err).Str("query", query).
				Msg("Failed to scan returned UUID")

			return nil, err
		}

		ids = append(ids, id)
	}

	if err = rows.Err(); err != nil {
		logger.Error().Err(err).Str("query", query).
			Msg("Error while reading returned UUIDs")

		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		logger.Error().Err(err).Str("query", query).
			Msg("Failed to commit transaction")

		return nil, err
	}

	return ids, nil
}

// -----------------------------------------------------------------------------
// Return integer ID
// -----------------------------------------------------------------------------

// ExecQueryReturnId executes a query and returns an integer ID.
func (db *Database) ExecQueryReturnId(
	ctx context.Context,
	query string,
	args ...any,
) (int, error) {

	logger := log.NewLogger()

	tx, err := db.Connection.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		logger.Error().Err(err).Str("query", query).
			Msg("Failed to start transaction")

		return -1, err
	}

	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		}
	}()

	var id int

	if err = tx.QueryRow(ctx, query, args...).Scan(&id); err != nil {
		logger.Error().Err(err).Str("query", query).
			Msg("Failed to execute query")

		return -1, err
	}

	if err = tx.Commit(ctx); err != nil {
		logger.Error().Err(err).Str("query", query).
			Msg("Failed to commit transaction")

		return -1, err
	}

	return id, nil
}

// -----------------------------------------------------------------------------
// Return int64 ID
// -----------------------------------------------------------------------------

// ExecQueryReturnBigId executes a query and returns an int64 ID.
func (db *Database) ExecQueryReturnBigId(
	ctx context.Context,
	query string,
	args ...any,
) (int64, error) {

	logger := log.NewLogger()

	tx, err := db.Connection.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		logger.Error().Err(err).Str("query", query).
			Msg("Failed to start transaction")

		return -1, err
	}

	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		}
	}()

	var id int64

	if err := tx.QueryRow(ctx, query, args...).Scan(&id); err != nil {
		logger.Error().Err(err).Str("query", query).
			Msg("Failed to execute query")

		return -1, err
	}

	if err := tx.Commit(ctx); err != nil {
		logger.Error().Err(err).Str("query", query).
			Msg("Failed to commit transaction")

		return -1, err
	}

	return id, nil
}

// -----------------------------------------------------------------------------
// Return UUID
// -----------------------------------------------------------------------------

// ExecQueryReturnUuid executes a query and returns a UUID.
func (db *Database) ExecQueryReturnUuid(
	ctx context.Context,
	query string,
	args ...any,
) (uuid.UUID, error) {

	logger := log.NewLogger()

	tx, err := db.Connection.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		logger.Error().Err(err).Str("query", query).
			Msg("Failed to start transaction")

		return uuid.Nil, err
	}

	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		}
	}()

	var id uuid.UUID

	if err := tx.QueryRow(ctx, query, args...).Scan(&id); err != nil {
		logger.Error().Err(err).Str("query", query).
			Msg("Failed to execute query")

		return uuid.Nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		logger.Error().Err(err).Str("query", query).
			Msg("Failed to commit transaction")

		return uuid.Nil, err
	}

	return id, nil
}

// -----------------------------------------------------------------------------
// Return UUID, including no-row handling
// -----------------------------------------------------------------------------

// ExecQueryNoRowsReturnUuid executes a query and returns a UUID.
// pgx.ErrNoRows is returned when the query has no result.
func (db *Database) ExecQueryNoRowsReturnUuid(
	ctx context.Context,
	query string,
	args ...any,
) (uuid.UUID, error) {

	logger := log.NewLogger()

	tx, err := db.Connection.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		logger.Error().Err(err).Str("query", query).
			Msg("Failed to start transaction")

		return uuid.Nil, err
	}

	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		}
	}()

	var id uuid.UUID

	err = tx.QueryRow(ctx, query, args...).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logger.Error().Err(err).Str("query", query).
				Msg("No results after query")
		} else {
			logger.Error().Err(err).Str("query", query).
				Msg("Failed to execute query")
		}

		return uuid.Nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		logger.Error().Err(err).Str("query", query).
			Msg("Failed to commit transaction")

		return uuid.Nil, err
	}

	return id, nil
}

// -----------------------------------------------------------------------------
// Count
// -----------------------------------------------------------------------------

// RunCountQuery executes a query returning a count.
func (db *Database) RunCountQuery(
	ctx context.Context,
	query string,
	args ...any,
) (int, error) {

	logger := log.NewLogger()

	var count int

	if err := db.Connection.QueryRow(ctx, query, args...).Scan(&count); err != nil {
		logger.Error().Err(err).Str("query", query).
			Msg("Failed to execute count query")

		return -1, err
	}

	return count, nil
}

// -----------------------------------------------------------------------------
// Exists
// -----------------------------------------------------------------------------

// CheckRowExists checks whether a query returns a boolean result.
func (db *Database) CheckRowExists(
	ctx context.Context,
	query string,
	args ...any,
) (bool, error) {

	logger := log.NewLogger()

	var rowExists bool

	err := db.Connection.QueryRow(ctx, query, args...).Scan(&rowExists)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}

		logger.Error().Err(err).Str("query", query).
			Msg("Failed to execute exists query")

		return false, err
	}

	return rowExists, nil
}

// -----------------------------------------------------------------------------
// Exists returning message
// -----------------------------------------------------------------------------

// CheckRowExistsReturnMsg executes a query returning a message.
func (db *Database) CheckRowExistsReturnMsg(
	ctx context.Context,
	query string,
	args ...any,
) (string, error) {

	logger := log.NewLogger()

	var resultMessage string

	if err := db.Connection.QueryRow(ctx, query, args...).Scan(&resultMessage); err != nil {
		logger.Error().Err(err).Str("query", query).
			Msg("Failed to execute query")

		return "", err
	}

	return resultMessage, nil
}
