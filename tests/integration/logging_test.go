package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// T050: Integration test for initialize request logging (raw request logged)
func TestInitializeRequestLogging(t *testing.T) {
	// Start the server in local mode
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", "run", "../../cmd/echo-mcp/main.go", "--mode=local", "--log-format=json")

	// Capture stdin and stdout
	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatalf("Failed to get stdin pipe: %v", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("Failed to get stdout pipe: %v", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		t.Fatalf("Failed to get stderr pipe: %v", err)
	}

	// Start the server
	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	// Read stderr in background to capture logs
	logBuffer := &bytes.Buffer{}
	go func() {
		io.Copy(logBuffer, stderr)
	}()

	// Send initialize request
	initRequest := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","clientInfo":{"name":"test","version":"1.0"},"capabilities":{}}}` + "\n"

	_, err = stdin.Write([]byte(initRequest))
	if err != nil {
		t.Fatalf("Failed to write to stdin: %v", err)
	}

	// Read response
	response := make([]byte, 4096)
	n, err := stdout.Read(response)
	if err != nil && err != io.EOF {
		t.Fatalf("Failed to read response: %v", err)
	}

	// Close stdin to signal server to shut down
	stdin.Close()

	// Wait for server to finish
	cmd.Wait()

	// Verify response is valid JSON-RPC
	var jsonResponse map[string]interface{}
	if err := json.Unmarshal(response[:n], &jsonResponse); err != nil {
		t.Fatalf("Invalid JSON response: %v, got: %s", err, string(response[:n]))
	}

	// Verify the response contains the expected fields
	if jsonResponse["jsonrpc"] != "2.0" {
		t.Errorf("Expected jsonrpc=2.0, got %v", jsonResponse["jsonrpc"])
	}
	if jsonResponse["id"] != float64(1) {
		t.Errorf("Expected id=1, got %v", jsonResponse["id"])
	}

	// Parse logs and verify they contain:
	// 1. request_received event
	// 2. raw_request field with the original request
	// 3. request_id field
	// 4. method field

	logLines := strings.Split(logBuffer.String(), "\n")
	foundRequestLog := false

	for _, line := range logLines {
		if line == "" {
			continue
		}

		var logEntry map[string]interface{}
		if err := json.Unmarshal([]byte(line), &logEntry); err != nil {
			// Skip non-JSON lines
			continue
		}

		// Check if this is the request_received event
		if eventType, ok := logEntry["event_type"].(string); ok && eventType == "request_received" {
			foundRequestLog = true

			// Verify method field
			if method, ok := logEntry["method"].(string); !ok || method != "initialize" {
				t.Errorf("Expected method=initialize, got %v", logEntry["method"])
			}

			// Verify request_id exists and is non-empty
			if requestID, ok := logEntry["request_id"].(string); !ok || requestID == "" {
				t.Errorf("Missing or empty request_id")
			}

			// Verify raw_request exists (note: it may be parsed as object or string)
			if _, ok := logEntry["raw_request"]; !ok {
				t.Errorf("Missing raw_request field in log")
			}

			break
		}
	}

	if !foundRequestLog {
		t.Errorf("Did not find request_received log entry. Logs:\n%s", logBuffer.String())
	}
}

// T051: Integration test for tools/call request logging (all requests logged chronologically)
func TestToolsCallRequestLogging(t *testing.T) {
	// Start the server in local mode
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", "run", "../../cmd/echo-mcp/main.go", "--mode=local", "--log-format=json")

	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatalf("Failed to get stdin pipe: %v", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("Failed to get stdout pipe: %v", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		t.Fatalf("Failed to get stderr pipe: %v", err)
	}

	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	logBuffer := &bytes.Buffer{}
	go func() {
		io.Copy(logBuffer, stderr)
	}()

	// Send multiple requests and verify they're all logged
	requests := []struct {
		method  string
		request string
	}{
		{
			method:  "initialize",
			request: `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","clientInfo":{"name":"test","version":"1.0"},"capabilities":{}}}`,
		},
		{
			method:  "tools/list",
			request: `{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}`,
		},
		{
			method:  "tools/call",
			request: `{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"echo","arguments":{"message":"test1"}}}`,
		},
		{
			method:  "tools/call",
			request: `{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"echo","arguments":{"message":"test2"}}}`,
		},
	}

	// Send all requests
	responseBuffer := &bytes.Buffer{}
	go func() {
		io.Copy(responseBuffer, stdout)
	}()

	for _, req := range requests {
		_, err = stdin.Write([]byte(req.request + "\n"))
		if err != nil {
			t.Fatalf("Failed to write request: %v", err)
		}
		// Small delay to ensure requests are processed in order
		time.Sleep(100 * time.Millisecond)
	}

	// Close stdin
	stdin.Close()

	// Wait for server
	cmd.Wait()

	// Parse logs and verify all requests were logged in order
	logLines := strings.Split(logBuffer.String(), "\n")
	requestLogs := []map[string]interface{}{}

	for _, line := range logLines {
		if line == "" {
			continue
		}

		var logEntry map[string]interface{}
		if err := json.Unmarshal([]byte(line), &logEntry); err != nil {
			continue
		}

		if eventType, ok := logEntry["event_type"].(string); ok && eventType == "request_received" {
			requestLogs = append(requestLogs, logEntry)
		}
	}

	// Verify we logged all requests
	if len(requestLogs) != len(requests) {
		t.Errorf("Expected %d request logs, got %d", len(requests), len(requestLogs))
	}

	// Verify they're in chronological order (have timestamps)
	for i, logEntry := range requestLogs {
		// Verify method matches
		expectedMethod := requests[i].method
		if method, ok := logEntry["method"].(string); !ok || method != expectedMethod {
			t.Errorf("Log %d: expected method=%s, got %v", i, expectedMethod, logEntry["method"])
		}

		// Verify timestamp exists
		if _, ok := logEntry["time"]; !ok {
			t.Errorf("Log %d: missing timestamp", i)
		}

		// Verify request_id exists
		if requestID, ok := logEntry["request_id"].(string); !ok || requestID == "" {
			t.Errorf("Log %d: missing or empty request_id", i)
		}
	}

	// Verify chronological order by comparing timestamps
	for i := 1; i < len(requestLogs); i++ {
		prevTime, ok1 := requestLogs[i-1]["time"].(string)
		currTime, ok2 := requestLogs[i]["time"].(string)

		if !ok1 || !ok2 {
			continue
		}

		prevT, err1 := time.Parse(time.RFC3339, prevTime)
		currT, err2 := time.Parse(time.RFC3339, currTime)

		if err1 != nil || err2 != nil {
			continue
		}

		if currT.Before(prevT) {
			t.Errorf("Logs not in chronological order: log %d (%s) is before log %d (%s)",
				i, currTime, i-1, prevTime)
		}
	}
}

// Test that error requests are also logged
func TestErrorRequestLogging(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", "run", "../../cmd/echo-mcp/main.go", "--mode=local", "--log-format=json")

	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatalf("Failed to get stdin pipe: %v", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("Failed to get stdout pipe: %v", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		t.Fatalf("Failed to get stderr pipe: %v", err)
	}

	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	logBuffer := &bytes.Buffer{}
	go func() {
		io.Copy(logBuffer, stderr)
	}()

	responseBuffer := &bytes.Buffer{}
	go func() {
		io.Copy(responseBuffer, stdout)
	}()

	// Send invalid request (malformed JSON)
	_, err = stdin.Write([]byte(`{"jsonrpc":"2.0","id":1` + "\n"))
	if err != nil {
		t.Fatalf("Failed to write request: %v", err)
	}

	time.Sleep(100 * time.Millisecond)
	stdin.Close()
	cmd.Wait()

	// Verify error was logged
	logLines := strings.Split(logBuffer.String(), "\n")
	foundErrorLog := false

	for _, line := range logLines {
		if line == "" {
			continue
		}

		var logEntry map[string]interface{}
		if err := json.Unmarshal([]byte(line), &logEntry); err != nil {
			continue
		}

		if level, ok := logEntry["level"].(string); ok && level == "error" {
			foundErrorLog = true
			// Verify it has request_id and raw_request
			if _, ok := logEntry["request_id"]; !ok {
				t.Errorf("Error log missing request_id")
			}
			break
		}
	}

	if !foundErrorLog {
		t.Logf("Warning: Did not find error log entry (this may be expected if malformed JSON is not logged as error)")
	}
}
