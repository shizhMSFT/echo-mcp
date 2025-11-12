package server

import (
	"github.com/shizhMSFT/echo-mcp/internal/protocol"
)

// HandleEcho implements the echo tool logic
// It returns the exact message received without any transformation
func HandleEcho(message string) (*protocol.ToolCallResult, error) {
	return &protocol.ToolCallResult{
		Content: []protocol.ContentItem{
			{
				Type: "text",
				Text: message,
			},
		},
	}, nil
}
