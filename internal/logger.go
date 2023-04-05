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
	var logger zerolog.Logger

	zerolog.SetGlobalLevel(level)
	if console {
		logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
	}
	logger = log.With().Timestamp().Logger()

	if level == zerolog.DebugLevel || level == zerolog.TraceLevel {
		logger = logger.With().Caller().Stack().Logger()
	}
	return logger
}
