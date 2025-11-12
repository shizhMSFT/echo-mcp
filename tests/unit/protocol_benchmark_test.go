package unit

import (
	"encoding/json"
	"testing"

	"github.com/shizhMSFT/echo-mcp/internal/protocol"
)

// T083: Benchmark test for JSON marshaling
func BenchmarkJSONMarshaling(b *testing.B) {
	request := protocol.JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/call",
		Params:  json.RawMessage(`{"name":"echo","arguments":{"message":"test"}}`),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(request)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkJSONUnmarshaling(b *testing.B) {
	data := []byte(`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"echo","arguments":{"message":"test"}}}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var req protocol.JSONRPCRequest
		err := json.Unmarshal(data, &req)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark marshaling with various payload sizes
func BenchmarkJSONMarshaling_SmallPayload(b *testing.B) {
	response := protocol.BuildResponse(1, map[string]interface{}{
		"status": "ok",
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		json.Marshal(response)
	}
}

func BenchmarkJSONMarshaling_MediumPayload(b *testing.B) {
	// 1KB message
	message := string(make([]byte, 1024))
	response := protocol.BuildResponse(1, map[string]interface{}{
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": message,
			},
		},
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		json.Marshal(response)
	}
}

func BenchmarkJSONMarshaling_LargePayload(b *testing.B) {
	// 100KB message
	message := string(make([]byte, 100*1024))
	response := protocol.BuildResponse(1, map[string]interface{}{
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": message,
			},
		},
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		json.Marshal(response)
	}
}

// Benchmark error response marshaling
func BenchmarkJSONMarshaling_ErrorResponse(b *testing.B) {
	response := protocol.BuildErrorResponse(1, protocol.InvalidParams, "Invalid parameters", "Missing message field")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		json.Marshal(response)
	}
}

// Benchmark InitializeResult marshaling (complex structure)
func BenchmarkJSONMarshaling_InitializeResult(b *testing.B) {
	result := protocol.InitializeResult{
		ProtocolVersion: protocol.ProtocolVersion,
		Capabilities: protocol.ServerCapabilities{
			Tools: &protocol.ToolsCapability{},
		},
		ServerInfo: protocol.ServerInfo{
			Name:    protocol.ServerName,
			Version: "0.1.0",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		json.Marshal(result)
	}
}

// Benchmark parsing request
func BenchmarkParseRequest(b *testing.B) {
	rawRequest := []byte(`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"echo","arguments":{"message":"test"}}}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := protocol.ParseRequest(rawRequest)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark building response
func BenchmarkBuildResponse(b *testing.B) {
	result := map[string]interface{}{
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": "test message",
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		protocol.BuildResponse(1, result)
	}
}

// Benchmark building error response
func BenchmarkBuildErrorResponse(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		protocol.BuildErrorResponse(1, protocol.MethodNotFound, "Method not found: unknown", nil)
	}
}
