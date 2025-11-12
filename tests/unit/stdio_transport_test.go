package unit

import (
	"bufio"
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/shizhMSFT/echo-mcp/internal/protocol"
)

// TestLineDelimitedJSONParsing tests that stdio transport correctly parses line-delimited JSON messages
func TestLineDelimitedJSONParsing(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantLines   int
		wantMethods []string
		wantValid   []bool
	}{
		{
			name: "single valid message",
			input: `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}
`,
			wantLines:   1,
			wantMethods: []string{"initialize"},
			wantValid:   []bool{true},
		},
		{
			name: "multiple valid messages",
			input: `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}
{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}
{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"echo","arguments":{"message":"test"}}}
`,
			wantLines:   3,
			wantMethods: []string{"initialize", "tools/list", "tools/call"},
			wantValid:   []bool{true, true, true},
		},
		{
			name: "message with unicode",
			input: `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"echo","arguments":{"message":"Hello 世界 🌍"}}}
`,
			wantLines:   1,
			wantMethods: []string{"tools/call"},
			wantValid:   []bool{true},
		},
		{
			name: "message with escaped characters",
			input: `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"echo","arguments":{"message":"Tab:\tNewline:\nQuote:\""}}}
`,
			wantLines:   1,
			wantMethods: []string{"tools/call"},
			wantValid:   []bool{true},
		},
		{
			name: "empty lines should be handled",
			input: `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}

{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}
`,
			wantLines:   3, // Including empty line
			wantMethods: []string{"initialize", "", "tools/list"},
			wantValid:   []bool{true, false, true},
		},
		{
			name: "malformed JSON",
			input: `{"jsonrpc":"2.0","id":1,"method":"initialize"
`,
			wantLines:   1,
			wantMethods: []string{""},
			wantValid:   []bool{false},
		},
		{
			name: "mixed valid and invalid",
			input: `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}
invalid json
{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}
`,
			wantLines:   3,
			wantMethods: []string{"initialize", "", "tools/list"},
			wantValid:   []bool{true, false, true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			scanner := bufio.NewScanner(reader)

			lineCount := 0
			for scanner.Scan() {
				line := scanner.Bytes()

				// Skip empty lines in validation
				if len(line) == 0 {
					lineCount++
					continue
				}

				// Try to parse as JSON-RPC request
				var req protocol.JSONRPCRequest
				err := json.Unmarshal(line, &req)

				if lineCount < len(tt.wantValid) {
					if tt.wantValid[lineCount] {
						if err != nil {
							t.Errorf("Line %d: expected valid JSON, got error: %v", lineCount, err)
						}
						if lineCount < len(tt.wantMethods) && req.Method != tt.wantMethods[lineCount] {
							t.Errorf("Line %d: expected method %q, got %q", lineCount, tt.wantMethods[lineCount], req.Method)
						}
					} else {
						if err == nil && len(line) > 0 {
							t.Errorf("Line %d: expected invalid JSON, but parsing succeeded", lineCount)
						}
					}
				}

				lineCount++
			}

			if err := scanner.Err(); err != nil {
				t.Fatalf("Scanner error: %v", err)
			}

			if lineCount != tt.wantLines {
				t.Errorf("Expected %d lines, got %d", tt.wantLines, lineCount)
			}
		})
	}
}

// TestStdioMessageBoundaries tests that messages are correctly separated by newlines
func TestStdioMessageBoundaries(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantCount int
	}{
		{
			name:      "single message with newline",
			input:     `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}` + "\n",
			wantCount: 1,
		},
		{
			name:      "single message without newline",
			input:     `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`,
			wantCount: 1,
		},
		{
			name: "multiple messages",
			input: `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}
{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}
{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{}}
`,
			wantCount: 3,
		},
		{
			name: "messages with multiple newlines",
			input: `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}

{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}


{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{}}
`,
			wantCount: 6, // Including empty lines
		},
		{
			name:      "empty input",
			input:     "",
			wantCount: 0,
		},
		{
			name:      "only newlines",
			input:     "\n\n\n",
			wantCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			scanner := bufio.NewScanner(reader)

			count := 0
			for scanner.Scan() {
				count++
			}

			if err := scanner.Err(); err != nil {
				t.Fatalf("Scanner error: %v", err)
			}

			if count != tt.wantCount {
				t.Errorf("Expected %d messages, got %d", tt.wantCount, count)
			}
		})
	}
}

// TestStdioLargeMessage tests handling of large messages
func TestStdioLargeMessage(t *testing.T) {
	// Create a large message (1MB)
	largeMessage := strings.Repeat("x", 1024*1024)
	request := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "tools/call",
		"params": map[string]interface{}{
			"name": "echo",
			"arguments": map[string]interface{}{
				"message": largeMessage,
			},
		},
	}

	jsonBytes, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("Failed to marshal large request: %v", err)
	}

	input := string(jsonBytes) + "\n"
	reader := strings.NewReader(input)

	// bufio.Scanner has a default buffer size limit (64KB)
	// We need to increase it for large messages
	scanner := bufio.NewScanner(reader)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 10*1024*1024) // 10MB max token size

	if !scanner.Scan() {
		t.Fatalf("Failed to scan large message: %v", scanner.Err())
	}

	line := scanner.Bytes()
	var req protocol.JSONRPCRequest
	if err := json.Unmarshal(line, &req); err != nil {
		t.Fatalf("Failed to parse large message: %v", err)
	}

	if req.Method != "tools/call" {
		t.Errorf("Expected method 'tools/call', got %q", req.Method)
	}
}

// TestStdioJSONWhitespace tests that JSON whitespace is handled correctly
func TestStdioJSONWhitespace(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "compact JSON",
			input:   `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}` + "\n",
			wantErr: false,
		},
		{
			name:    "pretty-printed JSON on one line",
			input:   `{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {}}` + "\n",
			wantErr: false,
		},
		{
			name:    "JSON with tabs",
			input:   "{\"jsonrpc\":\t\"2.0\",\t\"id\":\t1,\t\"method\":\t\"initialize\",\t\"params\":\t{}}\n",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			scanner := bufio.NewScanner(reader)

			if !scanner.Scan() {
				t.Fatalf("Failed to scan input")
			}

			line := scanner.Bytes()
			var req protocol.JSONRPCRequest
			err := json.Unmarshal(line, &req)

			if tt.wantErr && err == nil {
				t.Error("Expected error, got none")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}
		})
	}
}

// TestStdioBinaryData tests handling of binary data encoded in JSON
func TestStdioBinaryData(t *testing.T) {
	tests := []struct {
		name    string
		message string
		wantErr bool
	}{
		{
			name:    "base64 encoded data",
			message: "SGVsbG8sIFdvcmxkIQ==",
			wantErr: false,
		},
		{
			name:    "null bytes in string (escaped)",
			message: "Hello\\u0000World",
			wantErr: false,
		},
		{
			name:    "control characters (escaped)",
			message: "Line1\\nLine2\\rLine3\\tTabbed",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      1,
				"method":  "tools/call",
				"params": map[string]interface{}{
					"name": "echo",
					"arguments": map[string]interface{}{
						"message": tt.message,
					},
				},
			}

			jsonBytes, err := json.Marshal(request)
			if err != nil {
				t.Fatalf("Failed to marshal request: %v", err)
			}

			input := string(jsonBytes) + "\n"
			reader := bytes.NewReader([]byte(input))
			scanner := bufio.NewScanner(reader)

			if !scanner.Scan() {
				t.Fatalf("Failed to scan input")
			}

			line := scanner.Bytes()
			var req protocol.JSONRPCRequest
			err = json.Unmarshal(line, &req)

			if tt.wantErr && err == nil {
				t.Error("Expected error, got none")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}
		})
	}
}
