package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// # New Logger
func init() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
}

// # Print Info
// NewLogger returns a new zerolog.Logger instance writing to stdout
func NewLogger() *zerolog.Logger {
	logger := log.Output(zerolog.ConsoleWriter{
		Out: os.Stdout, TimeFormat: time.RFC3339,
	})

	return &logger
}
