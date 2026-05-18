package proxy

import (
	"encoding/json"
	"testing"
)

func TestTransformRequest_ModelReplacement(t *testing.T) {
	input := []byte(`{"model":"claude-sonnet-4-6","messages":[{"role":"user","content":"hi"}]}`)
	output, err := TransformRequest(input, "deepseek-v4-flash")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]any
	json.Unmarshal(output, &result)
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

	var result map[string]any
	json.Unmarshal(output, &result)

	sys := result["system"].([]any)
	block := sys[0].(map[string]any)
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

	var result map[string]any
	json.Unmarshal(output, &result)

	msgs := result["messages"].([]any)
	content := msgs[0].(map[string]any)["content"].([]any)
	block := content[0].(map[string]any)
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

	var result map[string]any
	json.Unmarshal(output, &result)

	tc := result["tool_choice"].(map[string]any)
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

	var result map[string]any
	json.Unmarshal(output, &result)

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
}
