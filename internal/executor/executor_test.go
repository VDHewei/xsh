package executor

import (
	"testing"

	"github.com/VDHewei/xsh/internal/types"
)

func TestExecuteHTTPTask(t *testing.T) {
	task := &types.Task{
		Type: types.TaskTypeHTTP,
		Raw:  "[GET] http://httpbin.org/get",
		HTTP: &types.HTTPTask{
			Method: types.GET,
			URL:    "http://httpbin.org/get",
		},
	}

	results := ExecuteTasks([]*types.Task{task})
	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	t.Logf("HTTP task result: %s", results[0])
}

func TestExecuteWaitTask(t *testing.T) {
	task := &types.Task{
		Type: types.TaskTypeWait,
		Raw:  "@wait: 10min",
		Wait: &types.WaitTask{
			Duration: "10min",
		},
	}

	results := ExecuteTasks([]*types.Task{task})
	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	t.Logf("Wait task result: %s", results[0])
}

func TestExecuteAskTask(t *testing.T) {
	task := &types.Task{
		Type: types.TaskTypeAsk,
		Raw:  "@ask: 检查服务是否正常?",
		Ask: &types.AskTask{
			Prompt: "检查服务是否正常?",
		},
	}

	results := ExecuteTasks([]*types.Task{task})
	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}
	t.Logf("Ask task result: %s", results[0])
}

func TestExecuteCheckTask(t *testing.T) {
	task := &types.Task{
		Type: types.TaskTypeCheck,
		Raw:  "@check: 验证部署结果",
		Check: &types.CheckTask{
			Prompt: "验证部署结果",
		},
	}

	results := ExecuteTasks([]*types.Task{task})
	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}
	t.Logf("Check task result: %s", results[0])
}

func TestExecuteMultipleTasks(t *testing.T) {
	tasks := []*types.Task{
		{Type: types.TaskTypeHTTP, Raw: "[GET] http://example.com/health", HTTP: &types.HTTPTask{Method: types.GET, URL: "http://example.com/health"}},
		{Type: types.TaskTypeWait, Raw: "@wait: 5min", Wait: &types.WaitTask{Duration: "5min"}},
		{Type: types.TaskTypeAsk, Raw: "@ask: Continue?", Ask: &types.AskTask{Prompt: "Continue?"}},
	}

	results := ExecuteTasks(tasks)
	if len(results) != 3 {
		t.Fatalf("Expected 3 results, got %d", len(results))
	}

	for i, r := range results {
		t.Logf("Result[%d]: %s", i, r)
	}
}

func TestExecuteUnknownTask(t *testing.T) {
	task := &types.Task{
		Type: "unknown",
		Raw:  "unknown task",
	}

	results := ExecuteTasks([]*types.Task{task})
	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}
	if results[0] != "[SKIP] unknown task type: unknown" {
		t.Errorf("Unexpected result: %s", results[0])
	}
}

func TestNormalizeDuration(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"10min", "10m"},
		{"5min", "5m"},
		{"30s", "30s"},
		{"1h", "1h"},
		{"10 min", "10m"},
	}

	for _, tt := range tests {
		result := normalizeDuration(tt.input)
		if result != tt.expected {
			t.Errorf("normalizeDuration(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}
