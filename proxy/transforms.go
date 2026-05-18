package proxy

import (
	"encoding/json"
)

// TransformRequest modifies the request body for DeepSeek compatibility:
// - Replaces the model name
// - Strips cache_control from system and content blocks
// - Removes disable_parallel_tool_use from tool_choice
func TransformRequest(body []byte, targetModel string) ([]byte, error) {
	var req map[string]any
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, err
	}

	req["model"] = targetModel

	if system, ok := req["system"]; ok {
		req["system"] = stripCacheControl(system)
	}

	if messages, ok := req["messages"]; ok {
		req["messages"] = stripCacheControlFromMessages(messages)
	}

	if toolChoice, ok := req["tool_choice"]; ok {
		if tc, ok := toolChoice.(map[string]any); ok {
			delete(tc, "disable_parallel_tool_use")
			if len(tc) == 0 {
				delete(req, "tool_choice")
			}
		}
	}

	return json.Marshal(req)
}

func stripCacheControl(v any) any {
	switch val := v.(type) {
	case map[string]any:
		delete(val, "cache_control")
		return val
	case []any:
		for i, item := range val {
			val[i] = stripCacheControl(item)
		}
		return val
	default:
		return v
	}
}

func stripCacheControlFromMessages(v any) any {
	messages, ok := v.([]any)
	if !ok {
		return v
	}
	for _, msg := range messages {
		if m, ok := msg.(map[string]any); ok {
			if content, ok := m["content"]; ok {
				m["content"] = stripCacheControl(content)
			}
		}
	}
	return messages
}
