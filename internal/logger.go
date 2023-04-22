package internal

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/diode"
)

type LogOption func(*zerolog.Logger)

// NewLogger creates a new instance zerolog logger.
func NewLogger(level zerolog.Level, console, noblock bool) zerolog.Logger {
	var l zerolog.Logger

	zerolog.SetGlobalLevel(level)

	var w io.Writer
	if noblock {
		w = diode.NewWriter(os.Stdout, 1000, 10*time.Millisecond, func(missed int) {
			fmt.Printf("Logger Dropped %d messages", missed)
		})
	} else {
		w = os.Stdout
	}

	if console {
		l = zerolog.New(zerolog.ConsoleWriter{Out: w, TimeFormat: time.RFC3339})
	} else {
		l = zerolog.New(w)
	}
	l = l.With().Timestamp().Logger()

	if level == zerolog.DebugLevel || level == zerolog.TraceLevel {
		l = l.With().Caller().Stack().Logger()
	}
	return l
}
