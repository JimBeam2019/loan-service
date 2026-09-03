package postgres

import (
	// "database/sql"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DatabaseParams struct {
	Hostname string
	Port     uint32
	Name     string
	Timeout  time.Duration
}

type DatabaseCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Database struct {
	Connection *pgxpool.Pool
	Mtx        sync.RWMutex
	Params     DatabaseParams
}
