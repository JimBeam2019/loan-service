package postgres

import (
	"context"
	"fmt"
	log "loan-service/pkg/logger"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// # New Database
func NewDatabase(
	ctx context.Context,
	params DatabaseParams,
	credentials DatabaseCredentials,
) (*Database, error) {
	database := &Database{
		Connection: nil,
		Mtx:        sync.RWMutex{},
		Params:     params,
	}

	if err := database.Reconnect(ctx, credentials); err != nil {
		return nil, err
	}

	return database, nil
}

// # Reconnect
func (db *Database) Reconnect(ctx context.Context, credentials DatabaseCredentials) error {
	logger := log.NewLogger()

	ctx, cancelContextFunc := context.WithTimeout(ctx, db.Params.Timeout)
	defer cancelContextFunc()

	logger.Info().Str("database", db.Params.Name).
		Str("hostname", db.Params.Hostname).
		Uint32("port", db.Params.Port).
		Str("username", credentials.Username).Msg("Connecting to database")

	connectionString := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		credentials.Username,
		credentials.Password,
		db.Params.Hostname,
		db.Params.Port,
		db.Params.Name,
	)

	config, err := pgxpool.ParseConfig(connectionString)
	if err != nil {
		logger.Error().Err(err).Msg("Unable to open database connection")
		return err
	}

	config.MaxConns = 25
	config.MaxConnIdleTime = 25
	config.MaxConnLifetime = 5 * time.Minute

	var connection *pgxpool.Pool

	for {
		connection, err = pgxpool.NewWithConfig(ctx, config)
		if err != nil {
			logger.Error().Err(err).
				Msg("Unable to create database connection pool")
			return err
		}

		if err = connection.Ping(ctx); err == nil {
			break
		}

		// Close the failed pool before retrying.
		connection.Close()

		select {
		case <-time.After(500 * time.Millisecond):
			continue

		case <-ctx.Done():
			logger.Error().Err(err).
				Msg("Failed to successfully ping database before context timeout")
			return err
		}
	}

	db.closeReplaceConnection(connection)

	logger.Info().Str("database", db.Params.Name).Msg("connecting to database")

	return nil
}

// (Private)
//
// # Close Replace Connection
func (db *Database) closeReplaceConnection(newDbConn *pgxpool.Pool) {
	db.Mtx.Lock()
	defer db.Mtx.Unlock()

	if db.Connection != nil {
		db.Connection.Close()
	}

	db.Connection = newDbConn
}

// # Close
func (db *Database) Close() error {
	db.Mtx.Lock()
	defer db.Mtx.Unlock()

	if db.Connection != nil {
		db.Connection.Close()
		db.Connection = nil
	}

	return nil
}
