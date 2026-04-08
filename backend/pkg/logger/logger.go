package logger

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

/**
 * Package logger provides a global structured logger built on zerolog.
 * Why zerolog?
 *   - Zero-allocation JSON logging — critical for high-throughput APIs
 *   - Each log line is a valid JSON object parseable by Datadog / Loki / CloudWatch
 *   - Simple API with method chaining for structured fields
 * Usage anywhere in the codebase:
 * logger.Info().Str("user_id", uid).Msg("User logged in")
 * logger.Error().Err(err).Str("request_id", rid).Msg("DB query failed")
*/

/**
 * Init initialises the global zerolog logger.
 * Call this once in main() before starting the server.
 * 
 * logger.Init("development")   // pretty console output
 * logger.Init("production")    // JSON output
*/
func Init(env string) {
	// Set global time field format to RFC3339 for interoperability
	zerolog.TimeFieldFormat = time.RFC3339

	var output io.Writer

	if env == "production" {
		// In production: pure JSON to stdout — ready for log aggregators
		output = os.Stdout
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	} else {
		// In development: coloured, human-readable console output
		output = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: "15:04:05",
			// Colour the level field
			NoColor: false,
		}
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	// Replace the global logger
	log.Logger = zerolog.New(output).
		With().
		Timestamp().
		Caller(). // includes file:line in every log entry
		Logger()
}

/**
 * Convenience wrappers around the global logger.
 * These are thin pass-throughs so callers don't
 * need to import zerolog directly.
*/

// Debug returns a debug-level event builder.
func Debug() *zerolog.Event {
	return log.Debug()
}

// Info returns an info-level event builder.
func Info() *zerolog.Event {
	return log.Info()
}

// Warn returns a warn-level event builder.
func Warn() *zerolog.Event {
	return log.Warn()
}

// Error returns an error-level event builder.
func Error() *zerolog.Event {
	return log.Error()
}

// Fatal returns a fatal-level event builder.
// Calling .Msg() on this will call os.Exit(1).
func Fatal() *zerolog.Event {
	return log.Fatal()
}

/**
 * WithRequestID returns an event builder pre-seeded with the request_id field.
 * Use this inside middleware or handlers for correlated log lines.
 * 
 * Usage:
 * logger.WithRequestID(rid).Error().Err(err).Msg("failed to fetch movie")
*/
func WithRequestID(requestID string) zerolog.Logger {
	return log.With().Str("request_id", requestID).Logger()
}
