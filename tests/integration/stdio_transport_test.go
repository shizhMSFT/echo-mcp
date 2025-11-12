package integration

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/shizhMSFT/echo-mcp/internal/protocol"
	"github.com/shizhMSFT/echo-mcp/internal/server"
	"github.com/shizhMSFT/echo-mcp/internal/transport"
)

// TestStdioTransport_InitializeHandshake tests the initialize flow
func TestStdioTransport_InitializeHandshake(t *testing.T) {
	input := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","clientInfo":{"name":"test-client","version":"1.0.0"},"capabilities":{}}}
`
	output := &bytes.Buffer{}
	inputReader := strings.NewReader(input)

	cfg := server.DefaultConfig()
	srv := server.NewServer(cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	tr := transport.NewStdioTransport(srv, inputReader, output)

	// Run transport in goroutine
	done := make(chan error, 1)
	go func() {
		done <- tr.Run(ctx)
	}()

	// Wait for processing
	select {
	case err := <-done:
		if err != nil && err != io.EOF {
			t.Fatalf("Transport error: %v", err)
		}
	case <-ctx.Done():
		t.Fatal("Timeout waiting for response")
	}

	// Parse response
	scanner := bufio.NewScanner(output)
	if !scanner.Scan() {
		t.Fatal("No response received")
	}

	var resp protocol.JSONRPCResponse
	if err := json.Unmarshal(scanner.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if resp.Error != nil {
		t.Fatalf("Unexpected error: %v", resp.Error)
	}

	// Verify response structure
	resultData, _ := json.Marshal(resp.Result)
	var initResult protocol.InitializeResult
	if err := json.Unmarshal(resultData, &initResult); err != nil {
		t.Fatalf("Failed to parse init result: %v", err)
	}

	if initResult.ServerInfo.Name != protocol.ServerName {
		t.Errorf("Expected server name %s, got %s", protocol.ServerName, initResult.ServerInfo.Name)
	}
}

// TestStdioTransport_EchoRequestResponse tests echo tool call
func TestStdioTransport_EchoRequestResponse(t *testing.T) {
	input := `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"echo","arguments":{"message":"Hello, Stdio!"}}}
`
	output := &bytes.Buffer{}
	inputReader := strings.NewReader(input)

	cfg := server.DefaultConfig()
	srv := server.NewServer(cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	tr := transport.NewStdioTransport(srv, inputReader, output)

	done := make(chan error, 1)
	go func() {
		done <- tr.Run(ctx)
	}()

	select {
	case err := <-done:
		if err != nil && err != io.EOF {
			t.Fatalf("Transport error: %v", err)
		}
	case <-ctx.Done():
		t.Fatal("Timeout waiting for response")
	}

	// Parse response
	scanner := bufio.NewScanner(output)
	if !scanner.Scan() {
		t.Fatal("No response received")
	}

	var resp protocol.JSONRPCResponse
	if err := json.Unmarshal(scanner.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if resp.Error != nil {
		t.Fatalf("Unexpected error: %v", resp.Error)
	}

	// Verify echo result
	resultData, _ := json.Marshal(resp.Result)
	var toolResult protocol.ToolCallResult
	if err := json.Unmarshal(resultData, &toolResult); err != nil {
		t.Fatalf("Failed to parse tool result: %v", err)
	}

	if len(toolResult.Content) == 0 {
		t.Fatal("Expected content in result")
	}

	if toolResult.Content[0].Text != "Hello, Stdio!" {
		t.Errorf("Expected 'Hello, Stdio!', got '%s'", toolResult.Content[0].Text)
	}
}

// TestStdioTransport_SequentialRequests tests multiple sequential requests
func TestStdioTransport_SequentialRequests(t *testing.T) {
	input := `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"echo","arguments":{"message":"First"}}}
{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"echo","arguments":{"message":"Second"}}}
{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"echo","arguments":{"message":"Third"}}}
`
	output := &bytes.Buffer{}
	inputReader := strings.NewReader(input)

	cfg := server.DefaultConfig()
	srv := server.NewServer(cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	tr := transport.NewStdioTransport(srv, inputReader, output)

	done := make(chan error, 1)
	go func() {
		done <- tr.Run(ctx)
	}()

	select {
	case err := <-done:
		if err != nil && err != io.EOF {
			t.Fatalf("Transport error: %v", err)
		}
	case <-ctx.Done():
		t.Fatal("Timeout waiting for responses")
	}

	// Verify we got 3 responses
	scanner := bufio.NewScanner(output)
	count := 0
	expectedMessages := []string{"First", "Second", "Third"}

	for scanner.Scan() {
		var resp protocol.JSONRPCResponse
		if err := json.Unmarshal(scanner.Bytes(), &resp); err != nil {
			t.Fatalf("Failed to parse response %d: %v", count+1, err)
		}

		resultData, _ := json.Marshal(resp.Result)
		var toolResult protocol.ToolCallResult
		if err := json.Unmarshal(resultData, &toolResult); err != nil {
			t.Fatalf("Failed to parse tool result %d: %v", count+1, err)
		}

		if toolResult.Content[0].Text != expectedMessages[count] {
			t.Errorf("Response %d: expected '%s', got '%s'",
				count+1, expectedMessages[count], toolResult.Content[0].Text)
		}
		count++
	}

	if count != 3 {
		t.Errorf("Expected 3 responses, got %d", count)
	}
}

// TestStdioTransport_GracefulShutdown tests graceful shutdown on stdin EOF
func TestStdioTransport_GracefulShutdown(t *testing.T) {
	input := `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"echo","arguments":{"message":"test"}}}
`
	output := &bytes.Buffer{}
	inputReader := strings.NewReader(input)

	cfg := server.DefaultConfig()
	srv := server.NewServer(cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	tr := transport.NewStdioTransport(srv, inputReader, output)

	err := tr.Run(ctx)
	// Should exit gracefully on EOF
	if err != nil && err != io.EOF {
		t.Fatalf("Expected graceful shutdown, got error: %v", err)
	}
}
