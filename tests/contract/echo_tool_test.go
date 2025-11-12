package contract

import (
	"encoding/json"
	"testing"

	"github.com/shizhMSFT/echo-mcp/internal/protocol"
	"github.com/shizhMSFT/echo-mcp/internal/server"
)

// TestToolsList_EchoToolSchema verifies the echo tool is listed with correct schema
func TestToolsList_EchoToolSchema(t *testing.T) {
	result := protocol.HandleToolsList()

	if len(result.Tools) != 1 {
		t.Fatalf("Expected 1 tool, got %d", len(result.Tools))
	}

	tool := result.Tools[0]

	// Verify tool name
	if tool.Name != "echo" {
		t.Errorf("Expected tool name 'echo', got '%s'", tool.Name)
	}

	// Verify description is non-empty
	if tool.Description == "" {
		t.Error("Expected non-empty description")
	}

	// Verify input schema
	if tool.InputSchema.Type != "object" {
		t.Errorf("Expected schema type 'object', got '%s'", tool.InputSchema.Type)
	}

	// Verify message property exists
	messageProp, ok := tool.InputSchema.Properties["message"]
	if !ok {
		t.Fatal("Expected 'message' property in schema")
	}

	if messageProp.Type != "string" {
		t.Errorf("Expected message type 'string', got '%s'", messageProp.Type)
	}

	// Verify message is required
	foundRequired := false
	for _, req := range tool.InputSchema.Required {
		if req == "message" {
			foundRequired = true
			break
		}
	}
	if !foundRequired {
		t.Error("Expected 'message' to be required")
	}
}

// TestToolsCall_SimpleString verifies echo with simple string
func TestToolsCall_SimpleString(t *testing.T) {
	params := protocol.ToolCallParams{
		Name: "echo",
		Arguments: map[string]interface{}{
			"message": "Hello, MCP!",
		},
	}

	paramsJSON, _ := json.Marshal(params)

	result, err := protocol.HandleToolsCall(paramsJSON, server.HandleEcho)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(result.Content) != 1 {
		t.Fatalf("Expected 1 content item, got %d", len(result.Content))
	}

	content := result.Content[0]
	if content.Type != "text" {
		t.Errorf("Expected content type 'text', got '%s'", content.Type)
	}

	if content.Text != "Hello, MCP!" {
		t.Errorf("Expected text 'Hello, MCP!', got '%s'", content.Text)
	}
}

// TestToolsCall_JSONString verifies JSON structure is preserved
func TestToolsCall_JSONString(t *testing.T) {
	jsonMessage := `{"key":"value","nested":{"data":123}}`

	params := protocol.ToolCallParams{
		Name: "echo",
		Arguments: map[string]interface{}{
			"message": jsonMessage,
		},
	}

	paramsJSON, _ := json.Marshal(params)

	result, err := protocol.HandleToolsCall(paramsJSON, server.HandleEcho)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.Content[0].Text != jsonMessage {
		t.Errorf("Expected JSON preserved, got '%s'", result.Content[0].Text)
	}
}

// TestToolsCall_EmptyString verifies empty string handling
func TestToolsCall_EmptyString(t *testing.T) {
	params := protocol.ToolCallParams{
		Name: "echo",
		Arguments: map[string]interface{}{
			"message": "",
		},
	}

	paramsJSON, _ := json.Marshal(params)

	result, err := protocol.HandleToolsCall(paramsJSON, server.HandleEcho)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.Content[0].Text != "" {
		t.Errorf("Expected empty string, got '%s'", result.Content[0].Text)
	}
}

// TestToolsCall_Unicode verifies unicode character handling
func TestToolsCall_Unicode(t *testing.T) {
	unicodeMessage := "Hello 世界 🌍"

	params := protocol.ToolCallParams{
		Name: "echo",
		Arguments: map[string]interface{}{
			"message": unicodeMessage,
		},
	}

	paramsJSON, _ := json.Marshal(params)

	result, err := protocol.HandleToolsCall(paramsJSON, server.HandleEcho)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.Content[0].Text != unicodeMessage {
		t.Errorf("Expected unicode preserved, got '%s'", result.Content[0].Text)
	}
}

// TestToolsCall_SpecialCharacters verifies special characters are preserved
func TestToolsCall_SpecialCharacters(t *testing.T) {
	specialMessage := "Line1\nLine2\tTabbed\"Quoted\""

	params := protocol.ToolCallParams{
		Name: "echo",
		Arguments: map[string]interface{}{
			"message": specialMessage,
		},
	}

	paramsJSON, _ := json.Marshal(params)

	result, err := protocol.HandleToolsCall(paramsJSON, server.HandleEcho)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.Content[0].Text != specialMessage {
		t.Errorf("Expected special chars preserved, got '%s'", result.Content[0].Text)
	}
}

// TestToolsCall_MissingMessage verifies error when message parameter is missing
func TestToolsCall_MissingMessage(t *testing.T) {
	params := protocol.ToolCallParams{
		Name:      "echo",
		Arguments: map[string]interface{}{},
	}

	paramsJSON, _ := json.Marshal(params)

	_, err := protocol.HandleToolsCall(paramsJSON, server.HandleEcho)
	if err == nil {
		t.Fatal("Expected error for missing message parameter")
	}

	rpcErr, ok := err.(*protocol.RPCError)
	if !ok {
		t.Fatalf("Expected RPCError, got %T", err)
	}

	if rpcErr.Code != protocol.InvalidParams {
		t.Errorf("Expected error code %d, got %d", protocol.InvalidParams, rpcErr.Code)
	}
}

// TestToolsCall_UnknownTool verifies error for unknown tool
func TestToolsCall_UnknownTool(t *testing.T) {
	params := protocol.ToolCallParams{
		Name: "unknown_tool",
		Arguments: map[string]interface{}{
			"message": "test",
		},
	}

	paramsJSON, _ := json.Marshal(params)

	_, err := protocol.HandleToolsCall(paramsJSON, server.HandleEcho)
	if err == nil {
		t.Fatal("Expected error for unknown tool")
	}

	rpcErr, ok := err.(*protocol.RPCError)
	if !ok {
		t.Fatalf("Expected RPCError, got %T", err)
	}

	if rpcErr.Code != protocol.MethodNotFound {
		t.Errorf("Expected error code %d, got %d", protocol.MethodNotFound, rpcErr.Code)
	}
}
