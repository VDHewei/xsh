package executor

import (
	"strings"
	"testing"

	"github.com/VDHewei/xsh/internal/types"
)

const mockBaseURL = "http://localhost:18080"

// --- HTTP Method Tests (通过测试) ---

func TestExecuteHTTP_GET(t *testing.T) {
	e := NewExecutor()
	task := httpTask(types.GET, mockBaseURL+"/get", nil, "")
	result := e.ExecuteTasks([]*types.Task{task})[0]
	if !result.Success {
		t.Fatalf("GET failed: %v", result.Error)
	}
	if !strings.Contains(result.Output, "success") {
		t.Errorf("unexpected output: %s", result.Output)
	}
	t.Logf("GET: %s", result.Output)
}

func TestExecuteHTTP_POST(t *testing.T) {
	e := NewExecutor()
	task := httpTask(types.POST, mockBaseURL+"/post",
		map[string]string{"X-Test": "post-test"}, `{"key":"value"}`)
	result := e.ExecuteTasks([]*types.Task{task})[0]
	if !result.Success {
		t.Fatalf("POST failed: %v", result.Error)
	}
	if !strings.Contains(result.Output, "created") {
		t.Errorf("unexpected output: %s", result.Output)
	}
	t.Logf("POST: %s", result.Output)
}

func TestExecuteHTTP_PUT(t *testing.T) {
	e := NewExecutor()
	task := httpTask(types.PUT, mockBaseURL+"/put", nil, `{"update":"data"}`)
	result := e.ExecuteTasks([]*types.Task{task})[0]
	if !result.Success {
		t.Fatalf("PUT failed: %v", result.Error)
	}
	if !strings.Contains(result.Output, "updated") {
		t.Errorf("unexpected output: %s", result.Output)
	}
	t.Logf("PUT: %s", result.Output)
}

func TestExecuteHTTP_PATCH(t *testing.T) {
	e := NewExecutor()
	task := httpTask(types.PATCH, mockBaseURL+"/patch", nil, `{"key":"value"}`)
	result := e.ExecuteTasks([]*types.Task{task})[0]
	if !result.Success {
		t.Fatalf("PATCH failed: %v", result.Error)
	}
	if !strings.Contains(result.Output, "patched") {
		t.Errorf("unexpected output: %s", result.Output)
	}
	t.Logf("PATCH: %s", result.Output)
}

func TestExecuteHTTP_DELETE(t *testing.T) {
	e := NewExecutor()
	task := httpTask(types.DELETE, mockBaseURL+"/delete", nil, "")
	result := e.ExecuteTasks([]*types.Task{task})[0]
	if !result.Success {
		t.Fatalf("DELETE failed: %v", result.Error)
	}
	if !strings.Contains(result.Output, "deleted") {
		t.Errorf("unexpected output: %s", result.Output)
	}
	t.Logf("DELETE: %s", result.Output)
}

func TestExecuteHTTP_HEAD(t *testing.T) {
	e := NewExecutor()
	task := httpTask("HEAD", mockBaseURL+"/head", nil, "")
	result := e.ExecuteTasks([]*types.Task{task})[0]
	if !result.Success {
		t.Fatalf("HEAD failed: %v", result.Error)
	}
	t.Logf("HEAD: %s", result.Output)
}

func TestExecuteHTTP_OPTIONS(t *testing.T) {
	e := NewExecutor()
	task := httpTask("OPTIONS", mockBaseURL+"/options", nil, "")
	result := e.ExecuteTasks([]*types.Task{task})[0]
	if !result.Success {
		t.Fatalf("OPTIONS failed: %v", result.Error)
	}
	t.Logf("OPTIONS: %s", result.Output)
}

// --- HTTP Error Scenario Tests (非通过测试) ---

func TestExecuteHTTP_Failure(t *testing.T) {
	e := NewExecutor()
	task := httpTask(types.GET, mockBaseURL+"/fail", nil, "")
	result := e.ExecuteTasks([]*types.Task{task})[0]
	if result.Success {
		t.Fatalf("expected failure but got success")
	}
	if !strings.Contains(result.Output, "500") {
		t.Errorf("expected 500 error: %s", result.Output)
	}
	t.Logf("FAIL: %s", result.Output)
}

func TestExecuteHTTP_Error(t *testing.T) {
	e := NewExecutor()
	task := httpTask(types.GET, mockBaseURL+"/error?code=502", nil, "")
	result := e.ExecuteTasks([]*types.Task{task})[0]
	if result.Success {
		t.Fatalf("expected error but got success")
	}
	if !strings.Contains(result.Output, "502") {
		t.Errorf("expected 502 error: %s", result.Output)
	}
	t.Logf("ERROR 502: %s", result.Output)
}

func TestExecuteHTTP_Echo(t *testing.T) {
	e := NewExecutor()
	task := httpTask(types.POST, mockBaseURL+"/echo",
		map[string]string{"X-Custom": "echo-val"}, `{"echo":"test"}`)
	result := e.ExecuteTasks([]*types.Task{task})[0]
	if !result.Success {
		t.Fatalf("Echo failed: %v", result.Error)
	}
	if !strings.Contains(result.Output, "echo") && !strings.Contains(result.Output, "200") {
		t.Errorf("unexpected output: %s", result.Output)
	}
	t.Logf("ECHO: %s", result.Output)
}

// --- Retry Test ---

func TestExecuteHTTP_Retry(t *testing.T) {
	e := NewExecutor()
	// Retry endpoint fails first 2 times, succeeds on 3rd
	task := httpTask(types.GET, mockBaseURL+"/retry", nil, "")
	result := e.ExecuteTasks([]*types.Task{task})[0]
	if !result.Success {
		t.Fatalf("Retry failed: %v", result.Error)
	}
	if !strings.Contains(result.Output, "success after retry") {
		t.Errorf("expected retry success: %s", result.Output)
	}
	t.Logf("RETRY: %s", result.Output)
}

// --- Connection Error Test ---

func TestExecuteHTTP_ConnectionRefused(t *testing.T) {
	e := NewExecutor()
	task := httpTask(types.GET, "http://localhost:19999/notexist", nil, "")
	result := e.ExecuteTasks([]*types.Task{task})[0]
	if result.Success {
		t.Fatalf("expected connection error but got success")
	}
	if !strings.Contains(result.Output, "HTTP_CONNECTION") && !strings.Contains(result.Output, "HTTP_ERROR") {
		t.Errorf("expected connection error: %s", result.Output)
	}
	t.Logf("Connection refused: %s", result.Output)
}

// --- Non-HTTP Tests ---

func TestExecuteWaitTask(t *testing.T) {
	e := NewExecutor()
	task := &types.Task{
		Type: types.TaskTypeWait,
		Raw:  "@wait: 10min",
		Wait: &types.WaitTask{Duration: "10min"},
	}
	result := e.ExecuteTasks([]*types.Task{task})[0]
	if !result.Success {
		t.Fatalf("Wait failed: %v", result.Error)
	}
	if !strings.Contains(result.Output, "10m") {
		t.Errorf("expected 10m in output: %s", result.Output)
	}
	t.Logf("Wait: %s", result.Output)
}

func TestExecuteAskTask(t *testing.T) {
	e := NewExecutor()
	task := &types.Task{
		Type: types.TaskTypeAsk,
		Raw:  "@ask: 检查服务是否正常?",
		Ask:  &types.AskTask{Prompt: "检查服务是否正常?"},
	}
	result := e.ExecuteTasks([]*types.Task{task})[0]
	if !result.Success {
		t.Fatalf("Ask failed: %v", result.Error)
	}
	t.Logf("Ask: %s", result.Output)
}

func TestExecuteCheckTask(t *testing.T) {
	e := NewExecutor()
	task := &types.Task{
		Type:  types.TaskTypeCheck,
		Raw:   "@check: 验证部署结果",
		Check: &types.CheckTask{Prompt: "验证部署结果"},
	}
	result := e.ExecuteTasks([]*types.Task{task})[0]
	if !result.Success {
		t.Fatalf("Check failed: %v", result.Error)
	}
	t.Logf("Check: %s", result.Output)
}

func TestExecuteUnknownTask(t *testing.T) {
	e := NewExecutor()
	task := &types.Task{
		Type: "unknown",
		Raw:  "unknown task",
	}
	result := e.ExecuteTasks([]*types.Task{task})[0]
	if result.Success {
		t.Fatalf("expected failure for unknown task")
	}
	t.Logf("Unknown: %s", result.Output)
}

func TestExecuteMultipleTasks(t *testing.T) {
	e := NewExecutor()
	tasks := []*types.Task{
		httpTask(types.GET, mockBaseURL+"/health", nil, ""),
		{Type: types.TaskTypeWait, Raw: "@wait: 5min", Wait: &types.WaitTask{Duration: "5min"}},
		{Type: types.TaskTypeAsk, Raw: "@ask: Continue?", Ask: &types.AskTask{Prompt: "Continue?"}},
	}
	results := e.ExecuteTasks(tasks)
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	for i, r := range results {
		t.Logf("Result[%d]: success=%v, output=%s", i, r.Success, r.Output)
	}
}

// --- RetryConfig Tests ---

func TestRetryConfig_Do_Success(t *testing.T) {
	cfg := NewRetryConfig(2, 0, 1.0) // 0 delay for fast test
	callCount := 0
	fn := func() (string, error) {
		callCount++
		if callCount < 3 {
			return "", &testError{msg: "temporary fail"}
		}
		return "success", nil
	}

	result, err := cfg.Do(fn)
	if err != nil {
		t.Fatalf("expected success after retries: %v", err)
	}
	if result != "success [retry:2]" {
		t.Errorf("expected 'success [retry:2]', got: %s", result)
	}
	t.Logf("Retry success after %d calls: %s", callCount, result)
}

func TestRetryConfig_Do_FailAll(t *testing.T) {
	cfg := NewRetryConfig(1, 0, 1.0)
	fn := func() (string, error) {
		return "", &testError{msg: "always fail"}
	}
	_, err := cfg.Do(fn)
	if err == nil {
		t.Fatal("expected error after all retries exhausted")
	}
	if !strings.Contains(err.Error(), "retries exhausted") {
		t.Errorf("unexpected error: %v", err)
	}
	t.Logf("Retry exhausted: %v", err)
}

// --- NormalizeDuration Tests ---

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

// --- Helpers ---

func httpTask(method types.HTTPMethod, url string, headers map[string]string, body string) *types.Task {
	return &types.Task{
		Type: types.TaskTypeHTTP,
		Raw:  "[" + string(method) + "] " + url,
		HTTP: &types.HTTPTask{
			Method:  method,
			URL:     url,
			Headers: headers,
			Body:    body,
		},
	}
}

type testError struct{ msg string }

func (e *testError) Error() string { return e.msg }
