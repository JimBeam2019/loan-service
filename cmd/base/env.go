package base

import "time"

type Env struct {
	UsePostgres      bool          `env:"USE_POSTGRES" long:"use-postgres"`
	PostgresHostname string        `env:"POSTGRES_SERVER_HOST" default:"127.0.0.1" long:"postgres-hostname"`
	PostgresPort     uint32        `env:"POSTGRES_SERVER_PORT" default:"5432" long:"postgres-port"`
	PostgresDBName   string        `env:"POSTGRES_DB" default:"loan-engine" long:"postgres-name"`
	PostgresUser     string        `env:"POSTGRES_USER" default:"loan-engine" long:"postgres-user"`
	PostgresPassword string        `env:"POSTGRES_PASSWORD" default:"password" long:"postgres-password"`
	PostgresTimeout  time.Duration `env:"POSTGRES_TIMEOUT" default:"10s" long:"postgres-timeout"`
}
