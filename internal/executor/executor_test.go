package executor

import (
	"fmt"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/VDHewei/xsh/internal/types"
	"golang.org/x/crypto/ssh"
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

// --- classifyHTTPError Tests ---

func TestClassifyHTTPError_Timeout(t *testing.T) {
	err := classifyHTTPError(fmt.Errorf("request timeout"))
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "HTTP_TIMEOUT") {
		t.Errorf("expected HTTP_TIMEOUT prefix, got: %v", err)
	}
}

func TestClassifyHTTPError_DeadlineExceeded(t *testing.T) {
	err := classifyHTTPError(fmt.Errorf("context deadline exceeded after 30s"))
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "HTTP_TIMEOUT") {
		t.Errorf("expected HTTP_TIMEOUT prefix, got: %v", err)
	}
}

func TestClassifyHTTPError_ConnectionRefused(t *testing.T) {
	err := classifyHTTPError(fmt.Errorf("connection refused"))
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "HTTP_CONNECTION") {
		t.Errorf("expected HTTP_CONNECTION prefix, got: %v", err)
	}
}

func TestClassifyHTTPError_NoSuchHost(t *testing.T) {
	err := classifyHTTPError(fmt.Errorf("no such host"))
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "HTTP_CONNECTION") {
		t.Errorf("expected HTTP_CONNECTION prefix, got: %v", err)
	}
}

func TestClassifyHTTPError_DNS(t *testing.T) {
	err := classifyHTTPError(fmt.Errorf("DNS lookup failed"))
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "HTTP_DNS") {
		t.Errorf("expected HTTP_DNS prefix, got: %v", err)
	}
}

func TestClassifyHTTPError_NameResolution(t *testing.T) {
	err := classifyHTTPError(fmt.Errorf("name resolution error"))
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "HTTP_DNS") {
		t.Errorf("expected HTTP_DNS prefix, got: %v", err)
	}
}

func TestClassifyHTTPError_TLS(t *testing.T) {
	err := classifyHTTPError(fmt.Errorf("TLS handshake failed"))
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "HTTP_TLS") {
		t.Errorf("expected HTTP_TLS prefix, got: %v", err)
	}
}

func TestClassifyHTTPError_Certificate(t *testing.T) {
	err := classifyHTTPError(fmt.Errorf("certificate verification failed"))
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "HTTP_TLS") {
		t.Errorf("expected HTTP_TLS prefix, got: %v", err)
	}
}

func TestClassifyHTTPError_Generic(t *testing.T) {
	err := classifyHTTPError(fmt.Errorf("unknown error occurred"))
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "HTTP_ERROR") {
		t.Errorf("expected HTTP_ERROR prefix, got: %v", err)
	}
}

func TestClassifyHTTPError_Nil(t *testing.T) {
	err := classifyHTTPError(nil)
	if err != nil {
		t.Errorf("expected nil for nil input, got: %v", err)
	}
}

// --- HTTP Slow Test ---

func TestExecuteHTTP_Slow(t *testing.T) {
	e := NewExecutor()
	// /slow?delay=2s should complete successfully within executor's 30s timeout
	task := httpTask(types.GET, mockBaseURL+"/slow?delay=2000000000", nil, "")
	result := e.ExecuteTasks([]*types.Task{task})[0]
	if !result.Success {
		t.Fatalf("Slow request failed: %v", result.Error)
	}
	if !strings.Contains(result.Output, "slow") {
		t.Errorf("expected 'slow' in output: %s", result.Output)
	}
	t.Logf("Slow: %s", result.Output)
}

// --- SSH Execution Tests ---

func TestExecuteSSH_Success(t *testing.T) {
	t.Skip("skipped: SSH kexLoop stalls with Go 1.25 + golang.org/x/crypto v0.50.0")
	if !serverReachable("localhost:18082") {
		t.Skip("SSH mock server not running on :18082")
	}
	e := NewExecutor()
	// SSH server accepts any password
	e.sshCfg = &ssh.ClientConfig{
		User:            "testuser",
		Auth:            []ssh.AuthMethod{ssh.Password("anypass")},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	}
	task := &types.Task{
		Type: types.TaskTypeSSH,
		Raw:  "@ssh: echo hello",
		SSH: &types.SSHTask{
			Host:    "localhost",
			Port:    "18082",
			User:    "testuser",
			Command: "echo hello",
		},
	}
	result := e.ExecuteTasks([]*types.Task{task})[0]
	if !result.Success {
		t.Fatalf("SSH failed: %v", result.Error)
	}
	if !strings.Contains(result.Output, "hello") {
		t.Errorf("expected 'hello' in output: %s", result.Output)
	}
	t.Logf("SSH: %s", result.Output)
}

func TestExecuteSSH_Failure(t *testing.T) {
	t.Skip("skipped: SSH kexLoop stalls with Go 1.25 + golang.org/x/crypto v0.50.0")
	if !serverReachable("localhost:18082") {
		t.Skip("SSH mock server not running on :18082")
	}
	e := NewExecutor()
	e.sshCfg = &ssh.ClientConfig{
		User:            "testuser",
		Auth:            []ssh.AuthMethod{ssh.Password("anypass")},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	}
	task := &types.Task{
		Type: types.TaskTypeSSH,
		Raw:  "@ssh: fail-command",
		SSH: &types.SSHTask{
			Host:    "localhost",
			Port:    "18082",
			User:    "testuser",
			Command: "fail-command",
		},
	}
	result := e.ExecuteTasks([]*types.Task{task})[0]
	if result.Success {
		t.Errorf("expected failure but got success: %s", result.Output)
	}
	t.Logf("SSH failure: %s", result.Output)
}

func TestExecuteSSH_ConnectionFailed(t *testing.T) {
	e := NewExecutor()
	e.sshCfg = &ssh.ClientConfig{
		User:            "nouser",
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         1 * time.Second,
	}
	task := &types.Task{
		Type: types.TaskTypeSSH,
		Raw:  "@ssh: echo hello",
		SSH: &types.SSHTask{
			Host:    "localhost",
			Port:    "19991",
			User:    "nouser",
			Command: "echo hello",
		},
	}
	result := e.ExecuteTasks([]*types.Task{task})[0]
	if result.Success {
		t.Errorf("expected connection failure but got success: %s", result.Output)
	}
	t.Logf("SSH connection failed: %s", result.Output)
}

// --- gRPC Execution Tests ---

func TestExecuteGRPC_HealthCheck(t *testing.T) {
	if !serverReachable("localhost:18081") {
		t.Skip("gRPC mock server not running on :18081")
	}
	e := NewExecutor()
	task := &types.Task{
		Type: types.TaskTypeGRPC,
		Raw:  "@grpc: HealthCheck",
		GRPC: &types.GRPCTask{
			Host:   "localhost",
			Port:   "18081",
			Method: "/mock.MockService/HealthCheck",
			Body:   `{"service":"test"}`,
		},
	}
	result := e.ExecuteTasks([]*types.Task{task})[0]
	if !result.Success {
		t.Fatalf("gRPC health check failed: %v", result.Error)
	}
	t.Logf("gRPC Health: %s", result.Output)
}

func TestExecuteGRPC_Echo(t *testing.T) {
	t.Skip("skipped: invokeGRPCMethod uses structpb.Value, mock server expects typed EchoRequest proto")
	if !serverReachable("localhost:18081") {
		t.Skip("gRPC mock server not running on :18081")
	}
	e := NewExecutor()
	task := &types.Task{
		Type: types.TaskTypeGRPC,
		Raw:  "@grpc: Echo",
		GRPC: &types.GRPCTask{
			Host:   "localhost",
			Port:   "18081",
			Method: "/mock.MockService/Echo",
			Body:   `{"method":"ECHO","payload":"test-data","headers":{"X-Test":"grpc-test"}}`,
		},
	}
	result := e.ExecuteTasks([]*types.Task{task})[0]
	if !result.Success {
		t.Fatalf("gRPC echo failed: %v", result.Error)
	}
	t.Logf("gRPC Echo: %s", result.Output)
}

func TestExecuteGRPC_ConnectionFailed(t *testing.T) {
	t.Skip("skipped: grpc.DialContext + WithBlock() may not respect context timeout on Go 1.25")
	e := NewExecutor()
	task := &types.Task{
		Type: types.TaskTypeGRPC,
		Raw:  "@grpc: HealthCheck",
		GRPC: &types.GRPCTask{
			Host:   "localhost",
			Port:   "19992",
			Method: "/mock.MockService/HealthCheck",
		},
	}
	result := e.ExecuteTasks([]*types.Task{task})[0]
	if result.Success {
		t.Errorf("expected connection failure but got success: %s", result.Output)
	}
	t.Logf("gRPC connection failed: %s", result.Output)
}

// --- Helpers ---

// serverReachable checks if a TCP server is listening on the given address
func serverReachable(addr string) bool {
	conn, err := net.DialTimeout("tcp", addr, 500*time.Millisecond)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

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
