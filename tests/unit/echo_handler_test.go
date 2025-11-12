package unit

import (
	"testing"

	"github.com/shizhMSFT/echo-mcp/internal/server"
)

// TestHandleEcho uses table-driven tests for message preservation
func TestHandleEcho(t *testing.T) {
	tests := []struct {
		name    string
		message string
	}{
		{
			name:    "simple string",
			message: "Hello, World!",
		},
		{
			name:    "empty string",
			message: "",
		},
		{
			name:    "unicode characters",
			message: "Hello 世界 🌍 مرحبا",
		},
		{
			name:    "special characters",
			message: "Line1\nLine2\tTab\"Quote\"",
		},
		{
			name:    "JSON string",
			message: `{"key":"value","number":123}`,
		},
		{
			name:    "very long string",
			message: string(make([]byte, 10000)),
		},
		{
			name:    "whitespace",
			message: "   spaces   ",
		},
		{
			name:    "numbers",
			message: "12345",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := server.HandleEcho(tt.message)

			if err != nil {
				t.Fatalf("Expected no error, got %v", err)
			}

			if result == nil {
				t.Fatal("Expected result, got nil")
			}

			if len(result.Content) != 1 {
				t.Fatalf("Expected 1 content item, got %d", len(result.Content))
			}

			content := result.Content[0]

			if content.Type != "text" {
				t.Errorf("Expected content type 'text', got '%s'", content.Type)
			}

			if content.Text != tt.message {
				t.Errorf("Message not preserved.\nExpected: %q\nGot: %q", tt.message, content.Text)
			}
		})
	}
}

// TestHandleEcho_IsNotError verifies the result is not marked as error
func TestHandleEcho_IsNotError(t *testing.T) {
	result, err := server.HandleEcho("test")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.IsError {
		t.Error("Expected IsError to be false")
	}
}
