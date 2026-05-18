package proxy

import (
	"encoding/json"
	"testing"
)

func mustUnmarshal(t *testing.T, data []byte) map[string]any {
	t.Helper()
	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}
	return result
}

func TestTransformRequest_ModelReplacement(t *testing.T) {
	input := []byte(`{"model":"claude-sonnet-4-6","messages":[{"role":"user","content":"hi"}]}`)
	output, err := TransformRequest(input, "deepseek-v4-flash")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := mustUnmarshal(t, output)
	if result["model"] != "deepseek-v4-flash" {
		t.Errorf("expected deepseek-v4-flash, got %v", result["model"])
	}
}

func TestTransformRequest_StripCacheControlFromSystem(t *testing.T) {
	input := []byte(`{
		"model":"claude-sonnet-4-6",
		"system":[{"type":"text","text":"be helpful","cache_control":{"type":"ephemeral"}}],
		"messages":[{"role":"user","content":"hi"}]
	}`)
	output, err := TransformRequest(input, "deepseek-v4-flash")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := mustUnmarshal(t, output)

	sys, ok := result["system"].([]any)
	if !ok {
		t.Fatal("system should be an array")
	}
	block, ok := sys[0].(map[string]any)
	if !ok {
		t.Fatal("system block should be an object")
	}
	if _, ok := block["cache_control"]; ok {
		t.Error("cache_control should be stripped from system blocks")
	}
}

func TestTransformRequest_StripCacheControlFromMessages(t *testing.T) {
	input := []byte(`{
		"model":"claude-sonnet-4-6",
		"messages":[
			{"role":"user","content":[{"type":"text","text":"hi","cache_control":{"type":"ephemeral"}}]}
		]
	}`)
	output, err := TransformRequest(input, "deepseek-v4-flash")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := mustUnmarshal(t, output)

	msgs, ok := result["messages"].([]any)
	if !ok {
		t.Fatal("messages should be an array")
	}
	content, ok := msgs[0].(map[string]any)["content"].([]any)
	if !ok {
		t.Fatal("content should be an array")
	}
	block, ok := content[0].(map[string]any)
	if !ok {
		t.Fatal("content block should be an object")
	}
	if _, ok := block["cache_control"]; ok {
		t.Error("cache_control should be stripped from content blocks")
	}
}

func TestTransformRequest_StripDisableParallelToolUse(t *testing.T) {
	input := []byte(`{
		"model":"claude-sonnet-4-6",
		"tool_choice":{"type":"auto","disable_parallel_tool_use":true},
		"messages":[{"role":"user","content":"hi"}]
	}`)
	output, err := TransformRequest(input, "deepseek-v4-flash")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := mustUnmarshal(t, output)

	tc, ok := result["tool_choice"].(map[string]any)
	if !ok {
		t.Fatal("tool_choice should exist and be an object")
	}
	if _, ok := tc["disable_parallel_tool_use"]; ok {
		t.Error("disable_parallel_tool_use should be stripped")
	}
}

func TestTransformRequest_PreservesOtherFields(t *testing.T) {
	input := []byte(`{
		"model":"claude-sonnet-4-6",
		"max_tokens":4096,
		"temperature":0.7,
		"stream":true,
		"system":"be helpful",
		"tools":[{"name":"bash","input_schema":{"type":"object"}}],
		"messages":[{"role":"user","content":"hi"}]
	}`)
	output, err := TransformRequest(input, "deepseek-v4-pro")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := mustUnmarshal(t, output)

	if result["model"] != "deepseek-v4-pro" {
		t.Errorf("model should be deepseek-v4-pro, got %v", result["model"])
	}
	if result["max_tokens"] != float64(4096) {
		t.Errorf("max_tokens should be preserved")
	}
	if result["stream"] != true {
		t.Errorf("stream should be preserved")
	}
	if result["temperature"] != 0.7 {
		t.Errorf("temperature should be preserved")
	}
	if result["system"] != "be helpful" {
		t.Errorf("system should be preserved")
	}
	tools, ok := result["tools"].([]any)
	if !ok {
		t.Fatal("tools should be an array")
	}
	if len(tools) != 1 {
		t.Errorf("tools should be preserved")
	}
}

func TestTransformRequest_InvalidJSON(t *testing.T) {
	_, err := TransformRequest([]byte(`not json`), "deepseek-v4-flash")
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestTransformRequest_SystemAsString(t *testing.T) {
	input := []byte(`{"model":"claude-sonnet-4-6","system":"be helpful","messages":[{"role":"user","content":"hi"}]}`)
	output, err := TransformRequest(input, "deepseek-v4-flash")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := mustUnmarshal(t, output)
	if result["system"] != "be helpful" {
		t.Errorf("expected system to be preserved as string, got %v", result["system"])
	}
}

func TestTransformRequest_ContentAsString(t *testing.T) {
	input := []byte(`{"model":"claude-sonnet-4-6","messages":[{"role":"user","content":"hi"}]}`)
	output, err := TransformRequest(input, "deepseek-v4-flash")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := mustUnmarshal(t, output)
	msgs, _ := result["messages"].([]any)
	msg := msgs[0].(map[string]any)
	if msg["content"] != "hi" {
		t.Errorf("expected string content to be preserved, got %v", msg["content"])
	}
}

func TestTransformRequest_EmptyToolChoice(t *testing.T) {
	input := []byte(`{
		"model":"claude-sonnet-4-6",
		"tool_choice":{"disable_parallel_tool_use":true},
		"messages":[{"role":"user","content":"hi"}]
	}`)
	output, err := TransformRequest(input, "deepseek-v4-flash")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := mustUnmarshal(t, output)
	if _, ok := result["tool_choice"]; ok {
		t.Error("tool_choice should be removed when it becomes empty")
	}
}
