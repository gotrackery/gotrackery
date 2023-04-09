package internal

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type LogOption func(*zerolog.Logger)

// NewLogger creates a new instance zerolog logger.
func NewLogger(level zerolog.Level, console bool) zerolog.Logger {
	var l zerolog.Logger

	zerolog.SetGlobalLevel(level)
	if console {
		l = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}) //nolint:staticcheck
	}
	l = log.With().Timestamp().Logger()

	if level == zerolog.DebugLevel || level == zerolog.TraceLevel {
		l = l.With().Caller().Stack().Logger()
	}
	return l
}
