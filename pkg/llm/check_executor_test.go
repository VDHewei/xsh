package llm

import (
	"strings"
	"testing"

	"github.com/VDHewei/xsh/internal/types"
)

func TestNewCheckExecutor(t *testing.T) {
	e := NewCheckExecutor(nil)
	if e == nil {
		t.Fatal("NewCheckExecutor should return non-nil executor")
	}
	if e.model != nil {
		t.Error("Expected nil model when created with nil")
	}
}

func TestCheckExecutorMockExecute(t *testing.T) {
	executor := NewCheckExecutor(nil)
	result, err := executor.Execute(&types.CheckTask{
		Prompt: "验证服务返回 200 OK",
	}, "GET http://localhost/health -> 200 OK")
	if err != nil {
		t.Fatalf("Execute should not fail: %v", err)
	}
	if result == nil {
		t.Fatal("Expected non-nil result")
	}
	if result.Prompt != "验证服务返回 200 OK" {
		t.Errorf("Expected prompt '验证服务返回 200 OK', got '%s'", result.Prompt)
	}
	if !result.Passed {
		t.Error("Mock execution should pass by default")
	}
	if !strings.Contains(result.Reason, "Mock") {
		t.Errorf("Mock reason should contain 'Mock', got '%s'", result.Reason)
	}
	if result.Context != "GET http://localhost/health -> 200 OK" {
		t.Errorf("Context should be preserved, got '%s'", result.Context)
	}
}

func TestCheckExecutorMockExecute_EmptyContext(t *testing.T) {
	executor := NewCheckExecutor(nil)
	result, err := executor.Execute(&types.CheckTask{
		Prompt: "检查迁移结果",
	}, "")
	if err != nil {
		t.Fatalf("Execute should not fail: %v", err)
	}
	if !result.Passed {
		t.Error("Mock execution should pass by default")
	}
	if result.Context != "" {
		t.Errorf("Context should be empty, got '%s'", result.Context)
	}
}

func TestCheckExecutorBuildCheckPrompt(t *testing.T) {
	executor := NewCheckExecutor(nil)
	task := &types.CheckTask{Prompt: "验证数据库迁移成功"}

	// With context
	prompt := executor.buildCheckPrompt(task, "前一步执行结果: OK")
	if prompt == "" {
		t.Error("buildCheckPrompt should return non-empty string")
	}
	if !strings.Contains(prompt, "验证数据库迁移成功") {
		t.Errorf("Prompt should contain task prompt, got: %s", prompt)
	}
	if !strings.Contains(prompt, "VERDICT: [PASS|FAIL]") {
		t.Error("Prompt should contain VERDICT format instruction")
	}
	if !strings.Contains(prompt, "前一步执行结果") {
		t.Error("Prompt should contain context when provided")
	}

	// Without context
	promptNoCtx := executor.buildCheckPrompt(task, "")
	if strings.Contains(promptNoCtx, "Prior execution results") {
		t.Error("Prompt without context should use default text")
	}
}

func TestParseCheckResult(t *testing.T) {
	tests := []struct {
		name           string
		response       string
		expectPassed   bool
		expectReason   string
	}{
		{
			name:           "pass verdict",
			response:       "VERDICT: PASS\nREASON: 所有检查通过，服务正常运行",
			expectPassed:   true,
			expectReason:   "所有检查通过，服务正常运行",
		},
		{
			name:           "fail verdict",
			response:       "VERDICT: FAIL\nREASON: 数据库连接超时",
			expectPassed:   false,
			expectReason:   "数据库连接超时",
		},
		{
			name:           "lowercase pass",
			response:       "verdict: pass\nreason: looks good",
			expectPassed:   true,
			expectReason:   "looks good",
		},
		{
			name:           "no reason line",
			response:       "VERDICT: PASS\nEverything is fine.",
			expectPassed:   true,
			expectReason:   "Everything is fine.",
		},
		{
			name:           "no verdict",
			response:       "Everything looks good here.",
			expectPassed:   false,
			expectReason:   "Everything looks good here.",
		},
		{
			name:           "empty response",
			response:       "",
			expectPassed:   false,
			expectReason:   "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			passed, reason := parseCheckResult(tc.response)
			if passed != tc.expectPassed {
				t.Errorf("parseCheckResult passed: expected %v, got %v", tc.expectPassed, passed)
			}
			if reason != tc.expectReason {
				t.Errorf("parseCheckResult reason: expected '%s', got '%s'", tc.expectReason, reason)
			}
		})
	}
}

func TestCheckExecutorWithUnloadedModel(t *testing.T) {
	// Model exists but not loaded => should fall back to mock
	model := NewModel("test-model")
	executor := NewCheckExecutor(model)

	result, err := executor.Execute(&types.CheckTask{
		Prompt: "验证迁移完成状态",
	}, "migration logs: success")
	if err != nil {
		t.Fatalf("Execute should not fail: %v", err)
	}
	if !result.Passed {
		t.Error("Mock should return passed=true")
	}
	if !strings.Contains(result.Reason, "Mock") {
		t.Errorf("Unloaded model should fall back to mock, got: %s", result.Reason)
	}
}
