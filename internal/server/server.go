package server

import (
	"os"
	"strconv"
)

// Mode represents the transport mode for the server
type Mode string

const (
	ModeLocal  Mode = "local"  // Stdio transport
	ModeRemote Mode = "remote" // HTTP transport
)

// Config holds the server configuration
type Config struct {
	Mode      Mode      // Transport mode (local or remote)
	Port      int       // HTTP port (remote mode only)
	LogFormat LogFormat // Log output format (text or json)
	LogLevel  string    // Log level (debug, info, warn, error)
}

// DefaultConfig returns a config with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		Mode:      ModeLocal,
		Port:      8080,
		LogFormat: LogFormatText,
		LogLevel:  "info",
	}
}

// LoadConfigFromEnv loads configuration from environment variables
func LoadConfigFromEnv() *Config {
	cfg := DefaultConfig()

	// Load mode
	if mode := os.Getenv("ECHO_MCP_MODE"); mode != "" {
		cfg.Mode = Mode(mode)
	}

	// Load port
	if portStr := os.Getenv("ECHO_MCP_PORT"); portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil {
			cfg.Port = port
		}
	}

	// Load log format
	if format := os.Getenv("ECHO_MCP_LOG_FORMAT"); format != "" {
		cfg.LogFormat = LogFormat(format)
	}

	// Load log level
	if level := os.Getenv("ECHO_MCP_LOG_LEVEL"); level != "" {
		cfg.LogLevel = level
	}

	return cfg
}

// Server represents the MCP server instance
type Server struct {
	Config *Config
	Logger *Logger
}

// NewServer creates a new server instance with the given configuration
func NewServer(cfg *Config) *Server {
	logger := NewLogger(cfg.LogFormat, cfg.LogLevel)
	return &Server{
		Config: cfg,
		Logger: logger,
	}
}
