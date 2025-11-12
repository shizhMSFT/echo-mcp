package server

import (
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// LogFormat represents the output format for logs
type LogFormat string

const (
	LogFormatText LogFormat = "text"
	LogFormatJSON LogFormat = "json"
)

// Logger handles structured logging for the MCP server
type Logger struct {
	log zerolog.Logger
}

// NewLogger creates a new logger with the specified format and level
func NewLogger(format LogFormat, level string) *Logger {
	// Parse log level
	logLevel, err := zerolog.ParseLevel(level)
	if err != nil {
		logLevel = zerolog.InfoLevel
	}

	var zlog zerolog.Logger

	// Configure output format
	if format == LogFormatText {
		// Human-readable console output
		output := zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}
		zlog = zerolog.New(output).Level(logLevel).With().Timestamp().Logger()
	} else {
		// Structured JSON output
		zlog = zerolog.New(os.Stdout).Level(logLevel).With().Timestamp().Logger()
	}

	return &Logger{log: zlog}
}

// LogRequest logs an incoming request with raw data
func (l *Logger) LogRequest(method string, rawRequest []byte) string {
	requestID := uuid.New().String()

	l.log.Info().
		Str("request_id", requestID).
		Str("event_type", "request_received").
		Str("method", method).
		RawJSON("raw_request", rawRequest).
		Msg("Request received")

	return requestID
}

// LogToolInvocation logs a tool invocation
func (l *Logger) LogToolInvocation(requestID, toolName string, rawRequest []byte) {
	l.log.Info().
		Str("request_id", requestID).
		Str("event_type", "tool_invoked").
		Str("tool", toolName).
		RawJSON("raw_request", rawRequest).
		Msg("Tool invoked")
}

// LogError logs an error event
func (l *Logger) LogError(requestID, message string, err error, rawRequest []byte) {
	l.log.Error().
		Str("request_id", requestID).
		Str("event_type", "error").
		Err(err).
		Str("error_message", message).
		RawJSON("raw_request", rawRequest).
		Msg("Error processing request")
}

// LogResponse logs a successful response
func (l *Logger) LogResponse(requestID string, method string) {
	l.log.Info().
		Str("request_id", requestID).
		Str("event_type", "response_sent").
		Str("method", method).
		Msg("Response sent")
}

// Info logs an informational message
func (l *Logger) Info(message string, fields map[string]interface{}) {
	event := l.log.Info()
	for k, v := range fields {
		event = event.Interface(k, v)
	}
	event.Msg(message)
}

// Debug logs a debug message
func (l *Logger) Debug(message string, fields map[string]interface{}) {
	event := l.log.Debug()
	for k, v := range fields {
		event = event.Interface(k, v)
	}
	event.Msg(message)
}

// Error logs an error message
func (l *Logger) Error(message string, err error, fields map[string]interface{}) {
	event := l.log.Error().Err(err)
	for k, v := range fields {
		event = event.Interface(k, v)
	}
	event.Msg(message)
}
