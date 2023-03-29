package internal

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// NewLogger creates a new instance zerolog logger.
func NewLogger(level zerolog.Level) zerolog.Logger {
	zerolog.SetGlobalLevel(level)

	return log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}).
		With().Timestamp().Caller().Stack().Logger()
}
