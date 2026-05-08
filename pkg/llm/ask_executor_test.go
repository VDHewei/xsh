package llm

import (
	"strings"
	"testing"

	"github.com/VDHewei/xsh/internal/types"
)

func TestNewAskExecutor(t *testing.T) {
	e := NewAskExecutor(nil)
	if e == nil {
		t.Fatal("NewAskExecutor should return non-nil executor")
	}
	// nil model is allowed (falls back to mock)
	if e.model != nil {
		t.Error("Expected nil model when created with nil")
	}
}

func TestAskExecutorMockExecute(t *testing.T) {
	// No model => mock execution
	executor := NewAskExecutor(nil)
	result, err := executor.Execute(&types.AskTask{
		Prompt: "确认开始部署吗?",
	})
	if err != nil {
		t.Fatalf("Execute should not fail: %v", err)
	}
	if result == nil {
		t.Fatal("Expected non-nil result")
	}
	if result.Prompt != "确认开始部署吗?" {
		t.Errorf("Expected prompt '确认开始部署吗?', got '%s'", result.Prompt)
	}
	if !strings.Contains(result.Response, "Mock") {
		t.Errorf("Mock response should contain 'Mock', got '%s'", result.Response)
	}
	if result.Suggestion == "" {
		t.Error("Suggestion should not be empty")
	}
	t.Logf("Mock ask result - Suggestion: %s", result.Suggestion)
}

func TestAskExecutorBuildAskPrompt(t *testing.T) {
	executor := NewAskExecutor(nil)
	task := &types.AskTask{Prompt: "测试问题"}

	prompt := executor.buildAskPrompt(task)
	if prompt == "" {
		t.Error("buildAskPrompt should return non-empty string")
	}
	if !strings.Contains(prompt, "测试问题") {
		t.Errorf("Prompt should contain the task prompt, got: %s", prompt)
	}
	if !strings.Contains(prompt, "RECOMMENDATION") {
		t.Error("Prompt should contain RECOMMENDATION instruction")
	}
	if !strings.Contains(prompt, "CONTINUE") {
		t.Error("Prompt should contain CONTINUE option")
	}
	if !strings.Contains(prompt, "REVIEW") {
		t.Error("Prompt should contain REVIEW option")
	}
	if !strings.Contains(prompt, "SKIP") {
		t.Error("Prompt should contain SKIP option")
	}
	t.Logf("Ask prompt length: %d", len(prompt))
}

func TestExtractSuggestion(t *testing.T) {
	tests := []struct {
		name     string
		response string
		expected string
	}{
		{
			name:     "standard format",
			response: "RECOMMENDATION: CONTINUE\nREASON: safe operation\nSUGGESTION: 可以继续执行部署",
			expected: "可以继续执行部署",
		},
		{
			name:     "suggestion with prefix",
			response: "分析完成\nSUGGESTION: 请先备份数据库再执行迁移\n这是重要步骤",
			expected: "请先备份数据库再执行迁移",
		},
		{
			name:     "no suggestion prefix",
			response: "这是一条简单的回复\n没有特定格式",
			expected: "没有特定格式",
		},
		{
			name:     "empty response",
			response: "",
			expected: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := extractSuggestion(tc.response)
			if got != tc.expected {
				t.Errorf("extractSuggestion: expected '%s', got '%s'", tc.expected, got)
			}
		})
	}
}

func TestAskExecutorWithUnloadedModel(t *testing.T) {
	// Model exists but not loaded => should fall back to mock
	model := NewModel("test-model")
	executor := NewAskExecutor(model)

	result, err := executor.Execute(&types.AskTask{
		Prompt: "检查版本兼容性?",
	})
	if err != nil {
		t.Fatalf("Execute should not fail: %v", err)
	}
	if !strings.Contains(result.Response, "Mock") {
		t.Errorf("Unloaded model should fall back to mock, got: %s", result.Response)
	}
}

func TestAskExecutorMockExecute_NilModel(t *testing.T) {
	executor := NewAskExecutor(nil)

	result, err := executor.Execute(&types.AskTask{
		Prompt: "确认回滚数据库?",
	})
	if err != nil {
		t.Fatalf("Execute with nil model should not fail: %v", err)
	}
	if result == nil {
		t.Fatal("Expected non-nil result")
	}
	if result.Prompt != "确认回滚数据库?" {
		t.Errorf("Expected prompt preserved, got '%s'", result.Prompt)
	}
}
